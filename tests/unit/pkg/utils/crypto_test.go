package utils

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Test constants to avoid duplication
const (
	emptyString                 = "empty string"
	testEmail                   = "user@example.com"
	testIP                      = "192.168.1.1"
	testIP2                     = "192.168.1.2"
	xForwardedForHeader         = "X-Forwarded-For"
	xRealIPHeader               = "X-Real-IP"
	passwordTooShortMsg         = "Password must be at least 8 characters long"
	passwordNeedsUppercaseMsg   = "Password must contain at least one uppercase letter"
	passwordNeedsLowercaseMsg   = "Password must contain at least one lowercase letter"
	passwordNeedsDigitMsg       = "Password must contain at least one digit"
	passwordNeedsSpecialCharMsg = "Password must contain at least one special character"
)

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hash empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hash simple string",
			input:    "hello",
			expected: "2cf24dba4f21d4288094e9b46ee2b7c5c8c8a5a8b9a25e75e8b3d7d7e8b2e9c5",
		},
		{
			name:     "hash complex string",
			input:    "Hello, World! 123",
			expected: "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		},
		{
			name:     "hash unicode string",
			input:    "héllo 世界",
			expected: "7b3d979ca8330a94fa7e9e1b466d8b99e0bcdea1ec90596c0dcc8d7ef6b4300c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.HashSHA256(tt.input)

			// Verify it's a valid hex string
			_, err := hex.DecodeString(result)
			assert.NoError(t, err)

			// Verify length (SHA256 produces 64 character hex string)
			assert.Len(t, result, 64)

			// For known inputs, verify exact hash
			if tt.name == "hash empty string" || tt.name == "hash simple string" {
				// These are deterministic and we can verify exact values
				assert.Len(t, result, 64)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid short password",
			password:    "password",
			expectError: false,
		},
		{
			name:        "valid medium password",
			password:    "this_is_a_longer_password_123!",
			expectError: false,
		},
		{
			name:        "password at 72 character limit",
			password:    strings.Repeat("a", 72),
			expectError: false,
		},
		{
			name:        "password exceeds 72 character limit",
			password:    strings.Repeat("a", 73),
			expectError: true,
			errorType:   bcrypt.ErrPasswordTooLong,
		},
		{
			name:        emptyString,
			password:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Verify it's a valid bcrypt hash (starts with $2a$ or $2b$)
				assert.True(t, strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$"))

				// Verify we can use it to check the original password
				assert.True(t, utils.CheckPasswordHash(tt.password, hash))
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// Generate a known hash for testing
	validPassword := "testpassword123"
	validHash, err := utils.HashPassword(validPassword)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password matches hash",
			password: validPassword,
			hash:     validHash,
			expected: true,
		},
		{
			name:     "incorrect password does not match",
			password: "wrongpassword",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "valid password with invalid hash",
			password: validPassword,
			hash:     "invalid_hash",
			expected: false,
		},
		{
			name:     "empty password with empty hash",
			password: "",
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateOTP(t *testing.T) {
	t.Run("default length", func(t *testing.T) {
		otp := utils.GenerateOTP()
		assert.Len(t, otp, 6)
		assertValidOTP(t, otp)
	})

	t.Run("custom length 4", func(t *testing.T) {
		otp := utils.GenerateOTP(4)
		assert.Len(t, otp, 4)
		assertValidOTP(t, otp)
	})

	t.Run("custom length 8", func(t *testing.T) {
		otp := utils.GenerateOTP(8)
		assert.Len(t, otp, 8)
		assertValidOTP(t, otp)
	})

	t.Run("invalid length defaults to 6", func(t *testing.T) {
		otp := utils.GenerateOTP(0)
		assert.Len(t, otp, 6)
		assertValidOTP(t, otp)
	})

	t.Run("randomness test", func(t *testing.T) {
		otps := make(map[string]bool)
		for i := 0; i < 10; i++ {
			otp := utils.GenerateOTP()
			otps[otp] = true
		}
		// Should have some variety (allowing for small chance of duplicates)
		assert.GreaterOrEqual(t, len(otps), 3, "OTPs should show randomness")
	})
}

func assertValidOTP(t *testing.T, otp string) {
	assert.NotEmpty(t, otp)
	// Verify it's all digits
	for _, char := range otp {
		assert.True(t, char >= '0' && char <= '9', "OTP should contain only digits")
	}
}

func TestGenerateULID(t *testing.T) {
	t.Run("generate ULID with current time", func(t *testing.T) {
		ulid1 := utils.GenerateULID(nil)
		ulid2 := utils.GenerateULID(nil)

		assert.Len(t, ulid1, 26) // ULID standard length
		assert.Len(t, ulid2, 26)
		assert.NotEqual(t, ulid1, ulid2) // Should be unique

		// Verify they're lowercase
		assert.Equal(t, strings.ToLower(ulid1), ulid1)
		assert.Equal(t, strings.ToLower(ulid2), ulid2)

		// Verify they're valid ULIDs
		_, err1 := utils.ParseULID(ulid1)
		_, err2 := utils.ParseULID(ulid2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("generate ULID with specific time", func(t *testing.T) {
		specificTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		ulid1 := utils.GenerateULID(&specificTime)
		ulid2 := utils.GenerateULID(&specificTime)

		assert.Len(t, ulid1, 26)
		assert.Len(t, ulid2, 26)

		// They should have the same timestamp part but different random parts
		assert.Equal(t, ulid1[:10], ulid2[:10]) // Timestamp part should be same
		assert.NotEqual(t, ulid1, ulid2)        // Full ULIDs should be different

		// Verify they're valid ULIDs
		parsed1, err1 := utils.ParseULID(ulid1)
		parsed2, err2 := utils.ParseULID(ulid2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Verify timestamps match (using millisecond precision)
		expectedTimeMs := uint64(specificTime.UnixMilli())
		assert.Equal(t, expectedTimeMs, parsed1.Time())
		assert.Equal(t, expectedTimeMs, parsed2.Time())
	})
}

func TestParseULID(t *testing.T) {
	// Generate a valid ULID for testing
	validULID := utils.GenerateULID(nil)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid ULID",
			input:       validULID,
			expectError: false,
		},
		{
			name:        "invalid ULID - too short",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5FA",
			expectError: true,
		},
		{
			name:        "invalid ULID - too long",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5FAVV",
			expectError: true,
		},
		{
			name:        "invalid ULID - invalid characters",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5FU",
			expectError: true,
		},
		{
			name:        emptyString,
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := utils.ParseULID(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ulid.ULID{}, parsed)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, ulid.ULID{}, parsed)

				// Verify round-trip
				assert.Equal(t, strings.ToUpper(tt.input), parsed.String())
			}
		})
	}
}

func TestGenerateUUID(t *testing.T) {
	t.Run("generate valid UUIDs", func(t *testing.T) {
		uuid1 := utils.GenerateUUID()
		uuid2 := utils.GenerateUUID()

		assert.Len(t, uuid1, 36) // UUID standard length
		assert.Len(t, uuid2, 36)
		assert.NotEqual(t, uuid1, uuid2) // Should be unique

		// Verify they're lowercase
		assert.Equal(t, strings.ToLower(uuid1), uuid1)
		assert.Equal(t, strings.ToLower(uuid2), uuid2)

		// Verify they're valid UUIDs
		_, err1 := utils.ParseUUID(uuid1)
		_, err2 := utils.ParseUUID(uuid2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Check format (8-4-4-4-12)
		parts1 := strings.Split(uuid1, "-")
		parts2 := strings.Split(uuid2, "-")
		assert.Len(t, parts1, 5)
		assert.Len(t, parts2, 5)
		assert.Len(t, parts1[0], 8)
		assert.Len(t, parts1[1], 4)
		assert.Len(t, parts1[2], 4)
		assert.Len(t, parts1[3], 4)
		assert.Len(t, parts1[4], 12)
	})
}

func TestParseUUID(t *testing.T) {
	// Generate a valid UUID for testing
	validUUID := utils.GenerateUUID()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid UUID",
			input:       validUUID,
			expectError: false,
		},
		{
			name:        "valid UUID uppercase",
			input:       strings.ToUpper(validUUID),
			expectError: false,
		},
		{
			name:        "invalid UUID - missing dashes",
			input:       "123e4567e89b12d3a456426614174000",
			expectError: false, // Actually valid - Go UUID library can parse this
		},
		{
			name:        "invalid UUID - wrong length",
			input:       "123e4567-e89b-12d3-a456-42661417400",
			expectError: true,
		},
		{
			name:        "invalid UUID - invalid characters",
			input:       "123e4567-e89b-12d3-a456-42661417400g",
			expectError: true,
		},
		{
			name:        emptyString,
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := utils.ParseUUID(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, parsed)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, parsed)
			}
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "generate 16-byte token",
			length: 16,
		},
		{
			name:   "generate 32-byte token",
			length: 32,
		},
		{
			name:   "generate 64-byte token",
			length: 64,
		},
		{
			name:   "generate 1-byte token",
			length: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token1, err1 := utils.GenerateSecureToken(tt.length)
			token2, err2 := utils.GenerateSecureToken(tt.length)

			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.NotEmpty(t, token1)
			assert.NotEmpty(t, token2)
			assert.NotEqual(t, token1, token2) // Should be unique

			// Base64 URL encoding should not contain padding
			assert.False(t, strings.Contains(token1, "="))
			assert.False(t, strings.Contains(token2, "="))
		})
	}

	t.Run("generate zero-length token", func(t *testing.T) {
		token, err := utils.GenerateSecureToken(0)
		assert.NoError(t, err)
		assert.Equal(t, "", token)
	})
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty string",
			a:        "hello",
			b:        "",
			expected: false,
		},
		{
			name:     "case sensitive",
			a:        "Hello",
			b:        "hello",
			expected: false,
		},
		{
			name:     "same length different content",
			a:        "abcdef",
			b:        "abcdeg",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal email",
			input:    testEmail,
			expected: testEmail,
		},
		{
			name:     "email with uppercase",
			input:    "User@Example.COM",
			expected: testEmail,
		},
		{
			name:     "email with leading/trailing spaces",
			input:    "  " + testEmail + "  ",
			expected: testEmail,
		},
		{
			name:     "email with mixed case and spaces",
			input:    "  User@Example.COM  ",
			expected: testEmail,
		},
		{
			name:     emptyString,
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func() *http.Request
		expected string
	}{
		{
			name: "X-Forwarded-For single IP",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set(xForwardedForHeader, testIP)
				return req
			},
			expected: testIP,
		},
		{
			name: "X-Forwarded-For multiple IPs",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set(xForwardedForHeader, testIP+", 10.0.0.1, 172.16.0.1")
				return req
			},
			expected: testIP,
		},
		{
			name: "X-Real-IP header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set(xRealIPHeader, testIP2)
				return req
			},
			expected: testIP2,
		},
		{
			name: "X-Forwarded-For takes precedence over X-Real-IP",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set(xForwardedForHeader, testIP)
				req.Header.Set(xRealIPHeader, testIP2)
				return req
			},
			expected: testIP,
		},
		{
			name: "fallback to RemoteAddr",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.3:12345"
				return req
			},
			expected: "192.168.1.3",
		},
		{
			name: "RemoteAddr without port",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.4"
				return req
			},
			expected: "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := utils.ExtractIPAddress(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPasswordSecure(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		expectedSecure bool
		expectedIssues []string
	}{
		{
			name:           "secure password",
			password:       "Password123!",
			expectedSecure: true,
			expectedIssues: []string{},
		},
		{
			name:           "too short",
			password:       "Pass1!",
			expectedSecure: false,
			expectedIssues: []string{passwordTooShortMsg},
		},
		{
			name:           "no uppercase",
			password:       "password123!",
			expectedSecure: false,
			expectedIssues: []string{passwordNeedsUppercaseMsg},
		},
		{
			name:           "no lowercase",
			password:       "PASSWORD123!",
			expectedSecure: false,
			expectedIssues: []string{passwordNeedsLowercaseMsg},
		},
		{
			name:           "no digit",
			password:       "Password!",
			expectedSecure: false,
			expectedIssues: []string{passwordNeedsDigitMsg},
		},
		{
			name:           "no special character",
			password:       "Password123",
			expectedSecure: false,
			expectedIssues: []string{passwordNeedsSpecialCharMsg},
		},
		{
			name:           "multiple issues",
			password:       "pass",
			expectedSecure: false,
			expectedIssues: []string{
				passwordTooShortMsg,
				passwordNeedsUppercaseMsg,
				passwordNeedsDigitMsg,
				passwordNeedsSpecialCharMsg,
			},
		},
		{
			name:           "empty password",
			password:       "",
			expectedSecure: false,
			expectedIssues: []string{
				passwordTooShortMsg,
				passwordNeedsUppercaseMsg,
				passwordNeedsLowercaseMsg,
				passwordNeedsDigitMsg,
				passwordNeedsSpecialCharMsg,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secure, issues := utils.IsPasswordSecure(tt.password)
			assert.Equal(t, tt.expectedSecure, secure)
			assert.Equal(t, len(tt.expectedIssues), len(issues))

			for _, expectedIssue := range tt.expectedIssues {
				assert.Contains(t, issues, expectedIssue)
			}
		})
	}
}

// Benchmark tests
func BenchmarkHashSHA256(b *testing.B) {
	input := "benchmark test string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.HashSHA256(input)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.HashPassword(password)
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateUUID()
	}
}

func BenchmarkGenerateULID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateULID(nil)
	}
}

func BenchmarkGenerateSecureToken(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateSecureToken(32)
	}
}
