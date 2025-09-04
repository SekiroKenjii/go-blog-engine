package jwt

import (
	"testing"
	"time"

	pkgjwt "github.com/SekiroKenjii/go-blog-engine/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret      = "test-secret-key-for-jwt-testing"
	testUserID      = "test-user-123"
	tokenMalformed  = "token is malformed"
	issuerName      = "thuongvo.dev"
	invalidTokenMsg = "invalid token"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expiry         time.Time
		tokenType      pkgjwt.TokenType
		expectError    bool
		errorSubstring string
	}{
		{
			name:        "generates access token successfully",
			userID:      testUserID,
			expiry:      time.Now().Add(time.Hour),
			tokenType:   pkgjwt.AccessToken,
			expectError: false,
		},
		{
			name:        "generates refresh token successfully",
			userID:      testUserID,
			expiry:      time.Now().Add(time.Hour * 24),
			tokenType:   pkgjwt.RefreshToken,
			expectError: false,
		},
		{
			name:        "handles past expiry time",
			userID:      testUserID,
			expiry:      time.Now().Add(-time.Hour),
			tokenType:   pkgjwt.AccessToken,
			expectError: false,
		},
		{
			name:        "handles empty user ID",
			userID:      "",
			expiry:      time.Now().Add(time.Hour),
			tokenType:   pkgjwt.AccessToken,
			expectError: false,
		},
		{
			name:        "handles custom token type",
			userID:      testUserID,
			expiry:      time.Now().Add(time.Hour),
			tokenType:   pkgjwt.TokenType("custom"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := pkgjwt.GenerateToken(tt.userID, tt.expiry, tt.tokenType, []byte(testSecret))

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token structure by parsing it
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(testSecret), nil
				})
				require.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				// Verify claims
				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, tt.userID, claims["user_id"])
				assert.Equal(t, string(tt.tokenType), claims["type"])
				assert.Equal(t, issuerName, claims["iss"])

				// Verify timing
				exp, ok := claims["exp"].(float64)
				require.True(t, ok)
				iat, ok := claims["iat"].(float64)
				require.True(t, ok)

				expectedExp := float64(tt.expiry.Unix())
				assert.InDelta(t, expectedExp, exp, 2.0)
				assert.True(t, exp >= iat)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// Generate a valid token for testing
	validExpiry := time.Now().Add(time.Hour)
	validToken, err := pkgjwt.GenerateToken(testUserID, validExpiry, pkgjwt.AccessToken, []byte(testSecret))
	require.NoError(t, err)

	// Generate an expired token
	expiredExpiry := time.Now().Add(-time.Hour)
	expiredToken, err := pkgjwt.GenerateToken(testUserID, expiredExpiry, pkgjwt.AccessToken, []byte(testSecret))
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		secret         []byte
		expectError    bool
		errorSubstring string
		expectedClaims *pkgjwt.CustomClaims
	}{
		{
			name:        "parses valid token successfully",
			token:       validToken,
			secret:      []byte(testSecret),
			expectError: false,
			expectedClaims: &pkgjwt.CustomClaims{
				UserID:    testUserID,
				TokenType: pkgjwt.AccessToken,
			},
		},
		{
			name:           "fails with invalid secret",
			token:          validToken,
			secret:         []byte("wrong-secret"),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
		{
			name:           "fails with expired token",
			token:          expiredToken,
			secret:         []byte(testSecret),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
		{
			name:           "fails with malformed token",
			token:          "invalid.jwt.token",
			secret:         []byte(testSecret),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
		{
			name:           "fails with empty token",
			token:          "",
			secret:         []byte(testSecret),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
		{
			name:           "fails with empty secret",
			token:          validToken,
			secret:         []byte(""),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
		{
			name:           "fails with token missing parts",
			token:          "onlyonepart",
			secret:         []byte(testSecret),
			expectError:    true,
			errorSubstring: invalidTokenMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := pkgjwt.ParseToken(tt.token, tt.secret)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.expectedClaims.UserID, claims.UserID)
				assert.Equal(t, tt.expectedClaims.TokenType, claims.TokenType)

				// Verify timing fields are set
				assert.NotNil(t, claims.IssuedAt)
				assert.NotNil(t, claims.ExpiresAt)
				assert.NotEmpty(t, claims.Issuer)
				assert.NotEmpty(t, claims.ID)
			}
		})
	}
}

func TestTokenTypeConstants(t *testing.T) {
	// Test that token type constants have expected values
	assert.Equal(t, pkgjwt.TokenType("access"), pkgjwt.AccessToken)
	assert.Equal(t, pkgjwt.TokenType("refresh"), pkgjwt.RefreshToken)
	assert.Equal(t, pkgjwt.TokenType("Bearer"), pkgjwt.DefaultAuthScheme)
}

func TestCustomClaimsValidation(t *testing.T) {
	t.Run("custom claims implements jwt.Claims interface", func(t *testing.T) {
		now := time.Now()
		claims := &pkgjwt.CustomClaims{
			UserID:    testUserID,
			TokenType: pkgjwt.AccessToken,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				Issuer:    issuerName,
			},
		}

		// Test that it implements the interface
		var _ jwt.Claims = claims

		// Test validation methods
		exp, err := claims.GetExpirationTime()
		assert.NoError(t, err)
		assert.NotNil(t, exp)
		assert.True(t, exp.After(now))

		iat, err := claims.GetIssuedAt()
		assert.NoError(t, err)
		assert.NotNil(t, iat)

		iss, err := claims.GetIssuer()
		assert.NoError(t, err)
		assert.Equal(t, issuerName, iss)
	})
}

func TestGenerateAndParseTokenRoundTrip(t *testing.T) {
	testCases := []struct {
		name      string
		userID    string
		tokenType pkgjwt.TokenType
		duration  time.Duration
	}{
		{
			name:      "access token round trip",
			userID:    "user1",
			tokenType: pkgjwt.AccessToken,
			duration:  time.Minute * 15,
		},
		{
			name:      "refresh token round trip",
			userID:    "user2",
			tokenType: pkgjwt.RefreshToken,
			duration:  time.Hour * 24,
		},
		{
			name:      "custom token round trip",
			userID:    "user3",
			tokenType: pkgjwt.TokenType("custom_type"),
			duration:  time.Hour * 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate token
			expiry := time.Now().Add(tc.duration)
			token, err := pkgjwt.GenerateToken(tc.userID, expiry, tc.tokenType, []byte(testSecret))
			require.NoError(t, err)
			require.NotEmpty(t, token)

			// Parse token
			claims, err := pkgjwt.ParseToken(token, []byte(testSecret))
			require.NoError(t, err)
			require.NotNil(t, claims)

			// Verify all fields
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.tokenType, claims.TokenType)
			assert.Equal(t, issuerName, claims.Issuer)
			assert.NotEmpty(t, claims.ID)

			// Verify timing
			assert.NotNil(t, claims.ExpiresAt)
			assert.NotNil(t, claims.IssuedAt)
			assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
		})
	}
}

func TestTokenWithSpecialCharacters(t *testing.T) {
	t.Run("handles special characters in user data", func(t *testing.T) {
		specialUserID := "user@example.com"

		expiry := time.Now().Add(time.Hour)
		token, err := pkgjwt.GenerateToken(specialUserID, expiry, pkgjwt.AccessToken, []byte(testSecret))
		require.NoError(t, err)

		claims, err := pkgjwt.ParseToken(token, []byte(testSecret))
		require.NoError(t, err)

		assert.Equal(t, specialUserID, claims.UserID)
	})
}

func TestTokenSecurityProperties(t *testing.T) {
	t.Run("different secrets produce different tokens", func(t *testing.T) {
		secret1 := []byte("secret1")
		secret2 := []byte("secret2")
		expiry := time.Now().Add(time.Hour)

		token1, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, secret1)
		require.NoError(t, err)

		token2, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, secret2)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)

		// token1 should not be parseable with secret2
		_, err = pkgjwt.ParseToken(token1, secret2)
		assert.Error(t, err)

		// token2 should not be parseable with secret1
		_, err = pkgjwt.ParseToken(token2, secret1)
		assert.Error(t, err)
	})

	t.Run("same parameters produce different tokens due to timing and UUID", func(t *testing.T) {
		expiry := time.Now().Add(time.Hour)

		token1, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, []byte(testSecret))
		require.NoError(t, err)

		// Small delay to ensure different issued at time
		time.Sleep(time.Millisecond)

		token2, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, []byte(testSecret))
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)

		// Both should be parseable
		claims1, err := pkgjwt.ParseToken(token1, []byte(testSecret))
		require.NoError(t, err)

		claims2, err := pkgjwt.ParseToken(token2, []byte(testSecret))
		require.NoError(t, err)

		// Claims should be the same except for timing and ID
		assert.Equal(t, claims1.UserID, claims2.UserID)
		assert.Equal(t, claims1.TokenType, claims2.TokenType)
		assert.NotEqual(t, claims1.ID, claims2.ID) // Different UUIDs
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("very long user ID", func(t *testing.T) {
		longUserID := string(make([]byte, 1000))
		for i := range longUserID {
			longUserID = string(rune('a' + i%26))
		}

		expiry := time.Now().Add(time.Hour)
		token, err := pkgjwt.GenerateToken(longUserID, expiry, pkgjwt.AccessToken, []byte(testSecret))
		require.NoError(t, err)

		claims, err := pkgjwt.ParseToken(token, []byte(testSecret))
		require.NoError(t, err)
		assert.Equal(t, longUserID, claims.UserID)
	})

	t.Run("very large secret", func(t *testing.T) {
		largeSecret := make([]byte, 1024)
		for i := range largeSecret {
			largeSecret[i] = byte(i % 256)
		}

		expiry := time.Now().Add(time.Hour)
		token, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, largeSecret)
		require.NoError(t, err)

		claims, err := pkgjwt.ParseToken(token, largeSecret)
		require.NoError(t, err)
		assert.Equal(t, testUserID, claims.UserID)
	})
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	expiry := time.Now().Add(time.Hour)
	secret := []byte(testSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseToken(b *testing.B) {
	expiry := time.Now().Add(time.Hour)
	secret := []byte(testSecret)
	token, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pkgjwt.ParseToken(token, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateAndParseToken(b *testing.B) {
	expiry := time.Now().Add(time.Hour)
	secret := []byte(testSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token, err := pkgjwt.GenerateToken(testUserID, expiry, pkgjwt.AccessToken, secret)
		if err != nil {
			b.Fatal(err)
		}

		_, err = pkgjwt.ParseToken(token, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}
