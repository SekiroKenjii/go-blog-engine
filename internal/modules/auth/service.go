package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/cache"
	"github.com/SekiroKenjii/go-blog-engine/internal/db"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"

	dbCtx "github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc/gen"
)

type AuthService struct {
	repo     *db.Repository
	tokenMgr ITokenManager
	cacheSvc abstract.ICacheService
	mailer   *mailers.Mailer
}

const (
	CryptographicOperationError = "Error occurred during a cryptographic operation"
	LogMessageFormat            = "%s: %v"
	EmailVerificationPrefix     = "email_verification:%s"
	PasswordResetPrefix         = "password_reset:%s"

	// Token expiration times
	EmailVerificationExpiry = 24 * time.Hour
	PasswordResetExpiry     = 1 * time.Hour
)

func NewAuthService() IAuthService {
	// Get the singleton mailer instance that was created during app bootstrap
	mailer := mailers.GetMailerInstance()
	if mailer == nil {
		panic("Mailer system not initialized. Ensure app.Bootstrap() is called before creating auth service")
	}

	return &AuthService{
		repo:     db.RepositoryInstance(),
		tokenMgr: TokenManagerInstance(),
		cacheSvc: cache.NewCacheService(),
		mailer:   mailer,
	}
}

// Register implements IAuthService.
func (s *AuthService) Register(ctx context.Context, email, firstName, lastName, password string) (string, response.ErrorCode) {
	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))

		return "", response.FATA000101
	}

	u, err := s.repo.CreateUser(ctx, dbCtx.CreateUserParams{
		ID:           utils.GenerateULID(nil),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: hashedPwd,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Create user operation failed: %v", err))

		return "", response.EBIZ000001
	}

	return u.ID, response.SBIZ000001
}

// Login implements IAuthService.
func (s *AuthService) Login(ctx context.Context, email, password, deviceID, ip, ua string) (*TokenPair, response.ErrorCode) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, response.EBIZ001000
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, response.EBIZ001001
	}

	err = s.repo.DeleteRefreshTokenByDevice(ctx, dbCtx.DeleteRefreshTokenByDeviceParams{
		UserID:   user.ID,
		DeviceID: deviceID,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete refresh token for user %s on device %s: %v", user.ID, deviceID, err))
	}

	token, err := s.tokenMgr.GenerateTokenPair(user.ID)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))

		return nil, response.FATA000101
	}

	tokenHash := utils.HashSHA256(token.RefreshToken)

	err = s.repo.StoreRefreshToken(ctx, dbCtx.StoreRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		DeviceID:  deviceID,
		Ip:        sql.NullString{String: ip, Valid: ip != ""},
		UserAgent: sql.NullString{String: ua, Valid: ua != ""},
		ExpiresAt: token.RefreshTokenExpires,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error occurred while storing refresh token: %v", err))

		return nil, response.FATA001001
	}

	return token, response.SBIZ000001
}

// RefreshToken implements IAuthService.
func (s *AuthService) RefreshToken(ctx context.Context, UserID, refreshToken string) (*TokenPair, response.ErrorCode) {
	tokenHash := utils.HashSHA256(refreshToken)

	dbToken, err := s.repo.GetRefreshToken(ctx, dbCtx.GetRefreshTokenParams{
		UserID:    UserID,
		TokenHash: tokenHash,
	})
	if err != nil {
		logger.Error("Refresh user token operation failed: invalid token")

		return nil, response.EBIZ001003
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		logger.Error("Refresh user token operation failed: token expired")

		err := s.repo.DeleteRefreshToken(ctx, dbCtx.DeleteRefreshTokenParams{
			UserID:    UserID,
			TokenHash: tokenHash,
		})

		if err != sql.ErrNoRows {
			logger.Error(fmt.Sprintf("Error deleting specific refresh token for user %s: %v", UserID, err))

			return nil, response.FATA001001
		}

		return nil, response.EBIZ001004
	}

	newAccessToken, newAccessTokenExpires, err := s.tokenMgr.GenerateAccessToken(UserID)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))

		return nil, response.FATA000101
	}

	return &TokenPair{
		AccessToken:         newAccessToken,
		AccessTokenExpires:  newAccessTokenExpires,
		RefreshToken:        refreshToken,
		RefreshTokenExpires: dbToken.ExpiresAt,
	}, response.SBIZ000001
}

// Logout implements IAuthService.
func (s *AuthService) Logout(ctx context.Context, userID, deviceID, refreshToken string) response.ErrorCode {
	tokenHash := utils.HashSHA256(refreshToken)

	// Try to delete by specific token first for better security
	if refreshToken != "" {
		err := s.repo.DeleteRefreshToken(ctx, dbCtx.DeleteRefreshTokenParams{
			UserID:    userID,
			TokenHash: tokenHash,
		})

		if err == nil {
			logger.Info(fmt.Sprintf("Successfully logged out user %s with specific token", userID))

			return response.SBIZ000001
		}

		if err != sql.ErrNoRows {
			logger.Error(fmt.Sprintf("Error deleting specific refresh token for user %s: %v", userID, err))

			return response.FATA001001
		}
	}

	// Fall back to device-based logout if specific token deletion failed or no token provided
	err := s.repo.DeleteRefreshTokenByDevice(ctx, dbCtx.DeleteRefreshTokenByDeviceParams{
		UserID:   userID,
		DeviceID: deviceID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info(fmt.Sprintf("No active session found for user %s on device %s", userID, deviceID))

			return response.SBIZ000001
		}

		logger.Error(fmt.Sprintf("Error logging out user %s from device %s: %v", userID, deviceID, err))

		return response.FATA001001
	}

	logger.Info(fmt.Sprintf("Successfully logged out user %s from device %s", userID, deviceID))

	return response.SBIZ000001
}

// VerifyEmail implements IAuthService.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) response.ErrorCode {
	userID, err := s.cacheSvc.Get(ctx, fmt.Sprintf(EmailVerificationPrefix, token))
	if err != nil || userID == "" {
		logger.Error(fmt.Sprintf("Failed to retrieve email verification token: %v", err))

		return response.EBIZ001005
	}

	// Try to delete the verification token from cache
	if err := s.cacheSvc.Delete(ctx, fmt.Sprintf(EmailVerificationPrefix, token)); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete email verification token from cache: %v", err))

		return response.FATA002001
	}

	if err := s.repo.MarkUserVerified(ctx, userID); err != nil {
		logger.Error(fmt.Sprintf("Failed to verify user %s: %v", userID, err))

		return response.FATA001001
	}

	// Send welcome email after successful verification (async)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err == nil {
		params := map[string]any{
			"firstName": user.FirstName,
		}
		_ = s.mailer.SendAsync(ctx, mailers.Strategies.Welcome(), user.Email, params)
	}

	return response.SBIZ000001
}

// SendVerificationEmail implements IAuthService.
func (s *AuthService) SendVerificationEmail(ctx context.Context, email, userID string) response.ErrorCode {
	// Generate a secure verification token using enhanced crypto utils
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))
		return response.FATA000101
	}

	// Store token in cache with expiration
	err = s.cacheSvc.Set(ctx, fmt.Sprintf(EmailVerificationPrefix, token), userID, int(EmailVerificationExpiry.Seconds()))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set email verification token in cache: %v", err))
		return response.FATA002001
	}

	// Get user details for email
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get user details for email: %v", err))
		return response.FATA001001
	}

	// Send verification email (async)
	params := map[string]any{
		"token":     token,
		"firstName": user.FirstName,
	}
	_ = s.mailer.SendAsync(ctx, mailers.Strategies.Verification(), user.Email, params)
	logger.Info(fmt.Sprintf("Verification email queued for sending to: %s", user.Email))

	return response.SBIZ000001
}

// SendPasswordResetEmail implements IAuthService.
func (s *AuthService) SendPasswordResetEmail(ctx context.Context, email string) response.ErrorCode {
	// First check if user exists
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// For security reasons, don't reveal if email exists or not
		logger.Info(fmt.Sprintf("Password reset requested for non-existent email: %s", email))
		return response.SBIZ000001
	}

	// Delete any existing password reset tokens for this user
	if err := s.repo.DeletePasswordResetTokensByUser(ctx, user.ID); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete existing password reset tokens: %v", err))
	}

	// Generate secure reset token using enhanced crypto utils
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))
		return response.FATA000101
	}

	tokenHash := utils.HashSHA256(token)

	// Store token in database with expiration
	expiresAt := time.Now().Add(PasswordResetExpiry)
	err = s.repo.StorePasswordResetToken(ctx, dbCtx.StorePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to store password reset token: %v", err))
		return response.FATA001001
	}

	// Send password reset email (async)
	params := map[string]any{
		"token":     token,
		"firstName": user.FirstName,
		"email":     user.Email,
	}
	_ = s.mailer.SendAsync(ctx, mailers.Strategies.PasswordReset(), user.Email, params)
	logger.Info(fmt.Sprintf("Password reset email queued for sending to: %s", user.Email))

	return response.SBIZ000001
}

// VerifyPasswordResetToken implements IAuthService.
func (s *AuthService) VerifyPasswordResetToken(ctx context.Context, email, token string) response.ErrorCode {
	// Verify user exists
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("Password reset token verification failed: user not found")
		return response.EBIZ001006
	}

	// Get and validate token
	tokenHash := utils.HashSHA256(token)
	dbToken, err := s.repo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		logger.Error("Password reset token verification failed: invalid token")
		return response.EBIZ001006
	}

	// Verify token belongs to the correct user
	if dbToken.UserID != user.ID {
		logger.Error("Password reset token verification failed: token user mismatch")
		return response.EBIZ001006
	}

	// Token is valid
	return response.SBIZ000001
}

// ResetPassword implements IAuthService.
func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword, token string) response.ErrorCode {
	// Verify user exists
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("Password reset failed: user not found")
		return response.EBIZ001006
	}

	// Get and validate token
	tokenHash := utils.HashSHA256(token)
	dbToken, err := s.repo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		logger.Error("Password reset failed: invalid token")
		return response.EBIZ001006
	}

	// Verify token belongs to the correct user
	if dbToken.UserID != user.ID {
		logger.Error("Password reset failed: token user mismatch")
		return response.EBIZ001006
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		logger.Error(fmt.Sprintf(LogMessageFormat, CryptographicOperationError, err))
		return response.FATA000101
	}

	// Update user password
	err = s.repo.UpdateUserPassword(ctx, dbCtx.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update user password: %v", err))
		return response.FATA001001
	}

	// Mark token as used
	if err := s.repo.MarkPasswordResetTokenUsed(ctx, tokenHash); err != nil {
		logger.Error(fmt.Sprintf("Failed to mark password reset token as used: %v", err))
		// Don't fail the operation, just log the error
	}

	// Delete all refresh tokens for this user to force re-login
	if err := s.repo.DeleteRefreshTokensByUser(ctx, user.ID); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete refresh tokens for user %s: %v", user.ID, err))
		// Don't fail the operation, just log the error
	}

	logger.Info(fmt.Sprintf("Password successfully reset for user %s", user.ID))
	return response.SBIZ000001
}
