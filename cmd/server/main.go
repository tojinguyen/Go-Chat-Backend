// @title           GoChat Backend API
// @version         1.0
// @description     A Real-time Chat Application Backend
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
package main

import (
	"context"
	"errors"
	"fmt"
	"gochat-backend/config"
	"gochat-backend/docs"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/router"
	"gochat-backend/internal/usecase"
	"gochat-backend/pkg/email"
	"gochat-backend/pkg/jwt"
	"gochat-backend/pkg/verification"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"net/http"

	cloudstorage "gochat-backend/internal/infra/cloudinaryinfra"
	"gochat-backend/internal/infra/mysqlinfra"
	"gochat-backend/internal/infra/redisinfra"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type App struct {
	config        *config.Environment
	logger        *logrus.Entry
	mysqlDatabase *mysqlinfra.Database
}

var listEnvSecret = []string{
	"Constants",
	"MysqlPassword",
	"RedisPassword",
	"AccessTokenSecretKey",
	"RefreshTokenSecretKey",
	"AwsSecretAccessKey",
}

func main() {
	logger := initLog()
	loggerStartServer := initStartServerLog()

	cfg := loadEnvironment()

	gin.SetMode(cfg.RunMode)

	loggerStartServer.Infof("System is running with %s mode", cfg.RunMode)

	if cfg.RunMode != "release" {
		if err := generateSwaggerDocs(loggerStartServer); err != nil {
			loggerStartServer.Warnf("Failed to generate Swagger documentation: %v", err)
			// Continue execution even if swagger generation fails
		}
	}

	// Initialize Database Connection
	db, err := InitDatabase(cfg)

	if err != nil {
		loggerStartServer.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Cloudinary Service
	cldService, err := InitCloudinary(cfg)
	if err != nil {
		loggerStartServer.Fatalf("Failed to create Cloudinary client: %v", err)
	}

	// Initialize Redis Service
	redisService, err := InitRedis(cfg)
	if err != nil {
		loggerStartServer.Fatalf("Failed to create Redis client: %v", err)
	}

	app := &App{
		config:        cfg,
		logger:        logger,
		mysqlDatabase: db,
	}

	// Initialize Services
	jwtService := jwt.NewJwtService(app.config, redisService)
	emailService := email.NewSMTPEmailService(app.config)
	verificationService := verification.NewVerificationService(app.config)

	// Initialize Repositories
	accountRepo := repository.NewAccountRepo(db, redisService)
	verificationRepo := repository.NewVerificationRepo(db)
	friendShipRepo := repository.NewFriendShipRepo(db)
	friendRequestRepo := repository.NewFriendRequestRepo(db)
	chatRoomRepo := repository.NewChatRoomRepo(db, redisService)
	messageRepo := repository.NewMessageRepo(db)
	statusRepo := repository.NewRedisStatusRepository(redisService)

	deps := &usecase.SharedDependencies{
		Config:              cfg,
		JwtService:          jwtService,
		EmailService:        emailService,
		VerificationService: verificationService,

		AccountRepo:              accountRepo,
		VerificationRegisterRepo: verificationRepo,
		FriendShipRepo:           friendShipRepo,
		FriendRequestRepo:        friendRequestRepo,
		ChatRoomRepo:             chatRoomRepo,
		MessageRepo:              messageRepo,
		StatusRepo:               statusRepo,

		CloudinaryStorage: cldService,
		RedisService:      redisService,
	}

	useCaseContainer := usecase.NewUseCaseContainer(deps)

	middleware := middleware.NewMiddleware(
		jwtService,
		logger,
		*app.config,
	)

	router := router.InitRouter(app.config, middleware, useCaseContainer, deps)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.Port),
		Handler: router,
	}

	if cfg.RunMode != "release" {
		setupSwagger()
	}

	done := make(chan bool)

	go func() {
		if err := GracefulShutDown(app.config, done, server); err != nil {
			loggerStartServer.Infof("Stop server shutdown error: %v", err.Error())
			return
		}
		loggerStartServer.Info("Stopped serving on Services")
	}()
	loggerStartServer.Infof("Start HTTP Server Successfully on PORT: %d", app.config.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		defer db.Close()
		loggerStartServer.Fatalf("Start HTTP Server Failed. Error: %s", err.Error())
	}
	<-done
	defer db.Close()
	loggerStartServer.Infof("Stopped backend application.")
}

func InitDatabase(cfg *config.Environment) (*mysqlinfra.Database, error) {
	db, err := mysqlinfra.ConnectMysql(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	database := mysqlinfra.NewMySqlDatabase(db)
	return database, nil
}

func InitCloudinary(cfg *config.Environment) (cloudstorage.CloudinaryService, error) {
	cldService, err := cloudstorage.NewCloudinaryService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary client: %v", err)
	}
	return cldService, nil
}

func InitRedis(cfg *config.Environment) (redisinfra.RedisService, error) {
	redisService, err := redisinfra.NewRedisService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}
	return redisService, nil
}

func GracefulShutDown(config *config.Environment, quit chan bool, server *http.Server) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-signals
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.SystemTimeOutSeconds)*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	close(quit)
	return nil
}

func setupSwagger() {
	docs.SwaggerInfo.Title = "GoChat Backend API"
	docs.SwaggerInfo.Description = "A Real-time Chat Application Backend."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	log.Println("=====================================================")
	log.Println("🚀 Server started successfully!")
	log.Println("📝 API Documentation: http://localhost:8080/docs/index.html")
	log.Println("=====================================================")
}

func initLog() *logrus.Entry {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		DisableQuote:    true,
		DisableColors:   true,
		FieldMap: logrus.FieldMap{
			"level": "logLevel",
		},
	})
	log := logrus.WithFields(logrus.Fields{
		"module": "backend",
	})
	return log
}

func initStartServerLog() *logrus.Entry {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		DisableQuote:    true,
		DisableColors:   true,
		DisableSorting:  true,
		FieldMap: logrus.FieldMap{
			"level": "logLevel",
		},
	})
	log := logrus.WithFields(logrus.Fields{
		"module":  "backend",
		"logType": "startServer",
	})
	return log
}

func loadEnvironment() *config.Environment {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load .env file: %v", err))
	}

	cfg, err := config.Load()

	if err != nil {
		logrus.Fatalf("Failed to load environment variables: %v", err)
	}

	fmt.Println("======================================================")

	v := reflect.ValueOf(cfg).Elem()
	for i := 0; i < v.NumField(); i++ {
		varName := v.Type().Field(i).Name
		varValue := v.Field(i).Interface()
		isLog := true
		for _, envSecret := range listEnvSecret {
			if varName == envSecret {
				isLog = false
				break
			}
		}
		if isLog {
			fmt.Printf("EnvKeyAndValue %s: '%v'\n", varName, varValue)
		}
	}

	fmt.Println("======================================================")

	return cfg
}

func generateSwaggerDocs(logger *logrus.Entry) error {
	logger.Info("Generating Swagger documentation...")

	// Check if swag is installed
	_, err := exec.LookPath("swag")
	if err != nil {
		return fmt.Errorf("swag command not found. Please install it with: go install github.com/swaggo/swag/cmd/swag@latest")
	}

	// Generate Swagger documentation
	cmd := exec.Command("swag", "init", "-g", "cmd/server/main.go", "-d", "./")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate Swagger documentation: %v", err)
	}

	logger.Info("Swagger documentation generated successfully!")
	return nil
}
