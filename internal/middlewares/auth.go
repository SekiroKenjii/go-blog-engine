package middlewares

import (
	"net/http"
	"strings"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/auth"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

// Auth creates a middleware to authenticate requests using JWT for Gin
func Auth() gin.HandlerFunc {
	tokenConfig := config.Instance().Security.Jwt

	return func(c *gin.Context) {
		// Extract token from the Authorization header
		tokenString := extractToken(c.Request)

		// Parse and validate the token
		claims, err := auth.ParseToken(tokenString, []byte(tokenConfig.SecretKey))
		if tokenString != "" && err != nil {
			response.AuthenticationHeaderError(c)

			return
		}

		if claims != nil {
			// Set claims in Gin context
			c.Set("claims", claims)
			c.Set("user_id", claims.UserID)
		}

		// Proceed to the next handler
		c.Next()
	}
}

// extractToken extracts the JWT token from the request's Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")

	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:7]) == "BEARER " {
		return bearerToken[7:]
	}

	return ""
}
