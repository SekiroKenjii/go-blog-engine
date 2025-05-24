package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/config"
)

var (
	tokenMgr     ITokenManager
	tokenMgrOnce sync.Once
)

type TokenManager struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	SecretKey  []byte
	RefreshKey []byte
}

type TokenPair struct {
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
}

func TokenManagerInstance() ITokenManager {
	tokenMgrOnce.Do(func() {
		tokenMgr = newTokenManager()
	})

	return tokenMgr
}

func newTokenManager() *TokenManager {
	jwtConf := config.Instance().Security.Jwt

	return &TokenManager{
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * 7 * time.Hour,
		SecretKey:  []byte(jwtConf.SecretKey),
		RefreshKey: []byte(jwtConf.RefreshKey),
	}
}

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

func (tm *TokenManager) GenerateAccessToken(userID string) (string, time.Time, error) {
	expires := time.Now().Add(tm.AccessTTL)
	token, err := generateToken(userID, expires, AccessToken, tm.SecretKey)

	return token, expires, err
}

func (tm *TokenManager) GenerateRefreshToken(length int) (string, time.Time, error) {
	expires := time.Now().Add(tm.RefreshTTL)
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(bytes)

	return token, expires, err
}

func (tm *TokenManager) ValidateAccessToken(token string) (*CustomClaims, error) {
	return ParseToken(token, tm.SecretKey)
}

func (tm *TokenManager) ValidateRefreshToken(token string) (*CustomClaims, error) {
	return ParseToken(token, tm.RefreshKey)
}
