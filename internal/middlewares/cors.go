package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	corsConfig := config.Instance().Security.Cors

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowedOrigin := "*"
		for o := range strings.SplitSeq(corsConfig.AllowOrigins, ",") {
			if o == "*" || o == origin {
				allowedOrigin = origin
				break
			}
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", corsConfig.AllowMethods)
		c.Header("Access-Control-Allow-Headers", corsConfig.AllowHeaders)
		c.Header("Access-Control-Max-Age", strconv.FormatInt(int64(corsConfig.MaxAge), 10))

		if corsConfig.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
