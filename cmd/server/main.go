package main

import (
	"crypto/tls"
	"fmt"
	"gochat-backend/docs"
	"gochat-backend/internal/config"
	"gochat-backend/internal/middleware"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

// @title           GoChat Backend API
// @version         1.0
// @description     A Real-time Chat Application Backend.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

var listEnvSecret = []string{
	"Constants",
	"MysqlPassword",
	"RedisPassword",
	"AccessTokenSecretKey",
	"RefreshTokenSecretKey",
	"AwsSecretAccessKey",
}

func main() {
	cfg := loadEnvironment()
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

	docs.SwaggerInfo.Title = "GoChat Backend API"
	docs.SwaggerInfo.Description = "A Real-time Chat Application Backend."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"https"}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	if cfg.RunMode != "debug" {
		r.Use(middleware.RedirectToHTTPS())
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist("localhost"),
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":80", certManager.HTTPHandler(nil))

	log.Println("Starting server on port:", port)

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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

	return cfg
}
