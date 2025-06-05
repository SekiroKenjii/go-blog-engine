package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/cache"
	"github.com/SekiroKenjii/go-blog-engine/internal/db"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"

	dbCtx "github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc/gen"
)

type AuthService struct {
	repo     *db.Repository
	tokenMgr ITokenManager
	cacheSvc abstract.ICacheService
}

func NewAuthService() IAuthService {
	return &AuthService{
		repo:     db.RepositoryInstance(),
		tokenMgr: TokenManagerInstance(),
		cacheSvc: cache.CacheServiceInstance(),
	}
}

// Register implements IAuthService.
func (s *AuthService) Register(ctx context.Context, email, firstName, lastName, password string) response.ErrorCode {
	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		logger.Error(fmt.Sprintf("Error occurred during a cryptographic operation: %v", err))

		return response.FATA000101
	}

	_, err = s.repo.CreateUser(ctx, dbCtx.CreateUserParams{
		ID:           utils.GenerateULID(nil),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: hashedPwd,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Create user operation failed: %v", err))

		return response.EBIZ000001
	}

	return response.SBIZ000001
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
		logger.Error(fmt.Sprintf("Error occurred during a cryptographic operation: %v", err))

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
		logger.Error(fmt.Sprintf("Error occurred during a cryptographic operation: %v", err))

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
	userID, err := s.cacheSvc.Get(ctx, fmt.Sprintf("email_verification:%s", token))
	if err != nil || userID == "" {
		logger.Error(fmt.Sprintf("Failed to retrieve email verification token: %v", err))

		return response.EBIZ000007
	}

	// Try to delete the verification token from cache
	if err := s.cacheSvc.Delete(ctx, fmt.Sprintf("email_verification:%s", token)); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete email verification token from cache: %v", err))

		return response.FATA002001
	}

	if err := s.repo.MarkUserVerified(ctx, userID); err != nil {
		logger.Error(fmt.Sprintf("Failed to verify user %s: %v", userID, err))

		return response.FATA001001
	}

	return response.SBIZ000001
}

// ResendVerificationEmail implements IAuthService.
func (s *AuthService) SendVerificationEmail(context.Context, string) response.ErrorCode {
	panic("unimplemented")
}

// SendPasswordResetEmail implements IAuthService.
func (s *AuthService) SendPasswordResetEmail(context.Context, string) response.ErrorCode {
	panic("unimplemented")
}

// VerifyPasswordResetToken implements IAuthService.
func (s *AuthService) VerifyPasswordResetToken(context.Context, string, string) response.ErrorCode {
	panic("unimplemented")
}

// ResetPassword implements IAuthService.
func (s *AuthService) ResetPassword(context.Context, string, string, string) response.ErrorCode {
	panic("unimplemented")
}
