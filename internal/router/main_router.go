package router

import (
	"gochat-backend/internal/config"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase"
	"gochat-backend/internal/validations"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	v1Router "gochat-backend/internal/router/v1"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(
	config *config.Environment,
	middleWare middleware.Middleware,
	useCaseContainer *usecase.UseCaseContainer,
) *gin.Engine {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(config.CorsAllowOrigins, ","),
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Access-Control-Allow-Headers",
			"Authorization",
			"X-XSRF-TOKEN",
			"screenId",
			"apiOrder",
		},
		ExposeHeaders: []string{
			"Content-Disposition",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(gin.Recovery())

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if v.RegisterValidation("customEmail", validations.CustomEmail) != nil {
			return nil
		}

		if v.RegisterValidation("customPassword", validations.CustomPassword) != nil {
			return nil
		}
	}

	apiRouter := router.Group("/api")

	apiRouter.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1Router.InitV1Router(
		apiRouter.Group("/v1", middleWare.RestLogger),
		middleWare,
		useCaseContainer,
	)

	return router
}
