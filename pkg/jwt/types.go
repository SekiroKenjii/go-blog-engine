package jwt

import "github.com/golang-jwt/jwt/v5"

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"type"`
	UserID    string    `json:"user_id"`
}
