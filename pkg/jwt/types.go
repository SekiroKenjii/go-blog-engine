package jwt

import "github.com/golang-jwt/jwt/v5"

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// CustomClaims defines the structure of the JWT claims used in the application.
// It includes standard JWT claims along with custom fields for token type and user ID.
type CustomClaims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"type"`
	UserID    string    `json:"user_id"`
}
