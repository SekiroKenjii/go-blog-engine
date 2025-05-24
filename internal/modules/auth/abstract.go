package auth

import (
	"context"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

type IAuthHandler interface {
	abstract.IHandler
	Register(*gin.Context)
	Login(*gin.Context)
	// RefreshToken(*gin.Context)
	// Logout(*gin.Context)
	// VerifyEmail(*gin.Context)
	// VerifyPhone(*gin.Context)
	// ResendVerificationEmail(*gin.Context)
	// ResendVerificationPhone(*gin.Context)
	// SendPasswordResetEmail(*gin.Context)
	// VerifyPasswordResetToken(*gin.Context)
	// ResetPassword(*gin.Context)
}

type IAuthService interface {
	Register(context.Context, RegisterRequest) response.ErrorCode
	Login(context.Context, LoginRequest, string, string, string) (*TokenPair, response.ErrorCode)
	RefreshToken(context.Context, string, string) (*TokenPair, response.ErrorCode)
}

type ITokenManager interface {
	GenerateTokenPair(string) (*TokenPair, error)
	GenerateAccessToken(string) (string, time.Time, error)
	GenerateRefreshToken(int) (string, time.Time, error)
	ValidateAccessToken(string) (*CustomClaims, error)
	ValidateRefreshToken(string) (*CustomClaims, error)
}
