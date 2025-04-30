package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToHTTPS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.TLS == nil {
			url := "https://" + ctx.Request.Host + ctx.Request.RequestURI
			ctx.Redirect(http.StatusMovedPermanently, url)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
