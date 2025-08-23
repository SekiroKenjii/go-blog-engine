package middlewares

import (
	"net/http"
	"strings"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/pkg/jwt"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

const BearerPrefix = "Bearer "

// RequireAuth creates a middleware that requires valid JWT authentication
func RequireAuth() gin.HandlerFunc {
	tokenConfig := config.Instance().Security.Jwt

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Failure(c, http.StatusUnauthorized, response.ESYS000011, []*response.ErrorInner{
				{Code: string(response.ESYS000011), Source: "Authorization header required"},
			}, nil)
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Failure(c, http.StatusUnauthorized, response.EBIZ001002, []*response.ErrorInner{
				{Code: string(response.EBIZ001002), Source: "Invalid authorization format"},
			}, nil)
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			response.Failure(c, http.StatusUnauthorized, response.EBIZ001002, []*response.ErrorInner{
				{Code: string(response.EBIZ001002), Source: "Token is required"},
			}, nil)
			c.Abort()
			return
		}

		// Parse and validate the token
		claims, err := jwt.ParseToken(token, []byte(tokenConfig.SecretKey))
		if err != nil {
			response.Failure(c, http.StatusUnauthorized, response.EBIZ001002, []*response.ErrorInner{
				{Code: string(response.EBIZ001002), Source: "Invalid or expired token"},
			}, nil)
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// RequireVerifiedEmail middleware checks if user's email is verified
func RequireVerifiedEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after RequireAuth
		userID, exists := c.Get("user_id")
		if !exists {
			response.Failure(c, http.StatusUnauthorized, response.EBIZ000003, []*response.ErrorInner{
				{Code: string(response.EBIZ000003), Source: "User not authenticated"},
			}, nil)
			c.Abort()
			return
		}

		// TODO: Check if user's email is verified in database
		// For now, we'll skip this check but keep the structure for future implementation
		_ = userID

		c.Next()
	}
}
