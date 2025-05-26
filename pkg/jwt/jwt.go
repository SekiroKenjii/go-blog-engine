package jwt

import (
	"errors"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a JWT token with the specified user ID, expiry time, token type, and signing key.
// It returns the signed token string or an error if the signing fails.
func GenerateToken(userID string, expiry time.Time, typ TokenType, key []byte) (string, error) {
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "thuongvo.dev",
			ID:        utils.GenerateUUID(),
		},
		TokenType: typ,
		UserID:    userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(key)
}

// ParseToken parses a JWT token string and returns the custom claims if the token is valid.
// It validates the signing method and extracts the claims, returning an error if the token is invalid or the claims cannot be parsed.
func ParseToken(tokenStr string, key []byte) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(key), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Additional validations could be added here
	return claims, nil
}
