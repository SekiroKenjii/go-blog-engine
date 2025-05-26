package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/pkg/jwt"
)

var (
	instance ITokenManager
	once     sync.Once
)

type TokenManager struct {
	accessTTL  time.Duration
	refreshTTL time.Duration
	secretKey  []byte
	refreshKey []byte
}

type TokenPair struct {
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
}

func TokenManagerInstance() ITokenManager {
	once.Do(func() {
		instance = newTokenManager()
	})

	return instance
}

func newTokenManager() *TokenManager {
	jwtConf := config.Instance().Security.Jwt

	return &TokenManager{
		accessTTL:  15 * time.Minute,
		refreshTTL: 24 * 7 * time.Hour,
		secretKey:  []byte(jwtConf.SecretKey),
		refreshKey: []byte(jwtConf.RefreshKey),
	}
}

// GenerateTokenPair implements ITokenManager.
func (tm *TokenManager) GenerateTokenPair(userID string) (*TokenPair, error) {
	access, accessExpires, err := tm.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refresh, refreshExpires, err := tm.GenerateRefreshToken(32)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:         access,
		AccessTokenExpires:  accessExpires,
		RefreshToken:        refresh,
		RefreshTokenExpires: refreshExpires,
	}, nil
}

// GenerateAccessToken implements ITokenManager.
func (tm *TokenManager) GenerateAccessToken(userID string) (string, time.Time, error) {
	expires := time.Now().Add(tm.accessTTL)
	token, err := jwt.GenerateToken(userID, expires, jwt.AccessToken, tm.secretKey)

	return token, expires, err
}

// GenerateRefreshToken implements ITokenManager.
func (tm *TokenManager) GenerateRefreshToken(length int) (string, time.Time, error) {
	expires := time.Now().Add(tm.refreshTTL)
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(bytes)

	return token, expires, err
}

// ValidateAccessToken implements ITokenManager.
func (tm *TokenManager) ValidateAccessToken(token string) (*jwt.CustomClaims, error) {
	return jwt.ParseToken(token, tm.secretKey)
}
