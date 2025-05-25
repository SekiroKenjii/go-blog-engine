package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc"
	"github.com/SekiroKenjii/go-blog-engine/internal/utils"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"

	dbCtx "github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc/gen"
)

type AuthService struct {
	Store        *sqlc.Store
	TokenManager ITokenManager
}

func NewService() IAuthService {
	return &AuthService{
		Store:        sqlc.Instance(),
		TokenManager: TokenManagerInstance(),
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) response.ErrorCode {
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error(fmt.Sprintf("Error occurred during a cryptographic operation: %v", err))

		return response.FATA000101
	}

	_, err = s.Store.CreateUser(ctx, dbCtx.CreateUserParams{
		ID:           utils.GenerateULID(nil),
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPwd,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Create user operation failed: %v", err))

		return response.EBIZ000001
	}

	return response.SBIZ000001
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest, deviceID, ip, ua string) (*TokenPair, response.ErrorCode) {
	user, err := s.Store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, response.EBIZ001000
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, response.EBIZ001001
	}

	err = s.Store.DeleteRefreshTokenByDevice(ctx, dbCtx.DeleteRefreshTokenByDeviceParams{
		UserID:   user.ID,
		DeviceID: deviceID,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete refresh token for user %s on device %s: %v", user.ID, deviceID, err))
	}

	token, err := s.TokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("Error occurred during a cryptographic operation: %v", err))

		return nil, response.FATA000101
	}

	tokenHash := utils.HashSHA256(token.RefreshToken)

	err = s.Store.StoreRefreshToken(ctx, dbCtx.StoreRefreshTokenParams{
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

func (s *AuthService) RefreshToken(ctx context.Context, UserID string, refreshToken string) (*TokenPair, response.ErrorCode) {
	dbToken, err := s.Store.GetRefreshToken(ctx, dbCtx.GetRefreshTokenParams{
		UserID:    UserID,
		TokenHash: utils.HashSHA256(refreshToken),
	})
	if err != nil {
		logger.Error("Refresh user token operation failed: invalid token")

		return nil, response.EBIZ001003
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		logger.Error("Refresh user token operation failed: token expired")

		_ = s.Store.DeleteRefreshToken(ctx, dbCtx.DeleteRefreshTokenParams{
			UserID:    UserID,
			TokenHash: refreshToken,
		})

		return nil, response.EBIZ001004
	}

	newAccessToken, newAccessTokenExpires, err := s.TokenManager.GenerateAccessToken(UserID)
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
