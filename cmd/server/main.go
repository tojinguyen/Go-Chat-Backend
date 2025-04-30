package main

import (
	"context"
	"errors"
	"fmt"
	"gochat-backend/docs"
	"gochat-backend/internal/config"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/router"
	"gochat-backend/internal/usecase/auth"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type App struct {
	config *config.Environment
	logger *logrus.Entry
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

	logger.Info("Starting GoChat Backend API...")

	cfg := loadEnvironment()

	gin.SetMode(cfg.RunMode)

	loggerStartServer.Infof("System is running with %s mode", cfg.RunMode)

	app := &App{
		config: cfg,
		logger: logger,
	}

	authUseCase := auth.NewAuthUseCase()

	middleware := middleware.NewMiddleware()

	router := router.InitRouter(app.config, middleware, authUseCase)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.Port),
		Handler: router,
	}

	setupSwagger()

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
		loggerStartServer.Fatalf("Start HTTP Server Failed. Error: %s", err.Error())
	}
	<-done
	loggerStartServer.Infof("Stopped backend application.")
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
	log.Println("📝 API Documentation: https://localhost:8080/api/v1/swagger/index.html")
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
