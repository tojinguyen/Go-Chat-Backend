package middleware

import (
	"github.com/gin-gonic/gin"
)

type Middleware interface {
	RestLogger(context *gin.Context)
	Authentication(context *gin.Context)
}

type middleware struct {
}

func NewMiddleware() Middleware {
	return &middleware{}
}
