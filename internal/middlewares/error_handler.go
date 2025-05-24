package middlewares

import (
	"net/http"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandler() gin.HandlerFunc {
	logger := logger.Instance()

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic occurred", zap.Any("error", err))
				c.Header("Content-Type", "application/json")

				response.Failure(c, http.StatusInternalServerError, response.FATA000001, nil, nil)
			}
		}()

		c.Next()
	}
}
