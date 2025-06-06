package auth

import (
	"context"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/pkg/jwt"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

type IAuthHandler interface {
	abstract.IHandler

	Register(*gin.Context)
	Login(*gin.Context)
	RefreshToken(*gin.Context)
	Logout(*gin.Context)
	VerifyEmail(*gin.Context)
	ResetPassword(*gin.Context)
}

type IAuthService interface {
	// Register allows a new user to create an account with their email, first name, last name, and password.
	// It hashes the password before storing it in the database.
	// Returns a user ID if registration is successful, or an error code if it fails.
	Register(context.Context, string, string, string, string) (string, response.ErrorCode)

	// Login allows a user to authenticate using their email and password, generating a token pair upon successful login.
	// It also deletes any existing refresh token for the user on the specified device before generating a new one.
	// It returns a TokenPair containing the access and refresh tokens, or an error code if the login fails.
	Login(context.Context, string, string, string, string, string) (*TokenPair, response.ErrorCode)

	// RefreshToken allows a user to obtain a new access token using a valid refresh token.
	// It validates the refresh token and generates a new token pair if valid.
	// Returns a TokenPair containing the new access and refresh tokens, or an error code if the refresh fails.
	RefreshToken(context.Context, string, string) (*TokenPair, response.ErrorCode)

	// Logout allows a user to log out from a specific device by deleting the refresh token associated with that device.
	// It returns an error code indicating the success or failure of the logout operation.
	Logout(context.Context, string, string, string) response.ErrorCode

	// VerifyEmail allows users to verify their email respectively.
	// It takes a context and the email verification token.
	// Returns an error code indicating the success or failure of the verification.
	VerifyEmail(context.Context, string) response.ErrorCode

	// SendVerificationEmail allows users to send a verification email.
	// It takes a context, the user's email address, and the user's ID.
	// Returns an error code indicating the success or failure of the operation.
	SendVerificationEmail(context.Context, string, string) response.ErrorCode

	// SendPasswordResetEmail allows users to initiate a password reset process by sending a reset email.
	// It takes a context and the user's email address.
	// Returns an error code indicating the success or failure of the operation.
	SendPasswordResetEmail(context.Context, string) response.ErrorCode

	// VerifyPasswordResetToken allows users to verify a password reset token.
	// It takes a context, the user's email, and the reset token.
	// Returns an error code indicating the success or failure of the verification.
	VerifyPasswordResetToken(context.Context, string, string) response.ErrorCode

	// ResetPassword allows users to reset their password using a valid reset token.
	// It takes a context, the user's email, the new password, and the reset token.
	// Returns an error code indicating the success or failure of the password reset operation.
	ResetPassword(context.Context, string, string, string) response.ErrorCode
}

type ITokenManager interface {
	// GenerateTokenPair generates a new token pair (access and refresh tokens) for a given user ID.
	// It returns a TokenPair containing the access token, refresh token, and their respective expiration times.
	GenerateTokenPair(string) (*TokenPair, error)

	// GenerateAccessToken generates a new access token for a given user ID.
	// It returns the access token string, its expiration time, and an error if any.
	GenerateAccessToken(string) (string, time.Time, error)

	// GenerateRefreshToken generates a new refresh token for a given user ID.
	// It returns the refresh token string, its expiration time, and an error if any.
	GenerateRefreshToken(int) (string, time.Time, error)

	// ValidateAccessToken validates the provided access token and returns the custom claims if valid.
	// If the token is invalid or expired, it returns an error.
	ValidateAccessToken(string) (*jwt.CustomClaims, error)
}
