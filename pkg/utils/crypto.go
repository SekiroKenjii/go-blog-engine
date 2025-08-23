package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

// HashSHA256 hashes a string using SHA-256 and returns the hexadecimal representation of the hash.
func HashSHA256(raw string) string {
	h := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(h[:])
}

// HashPassword hashes a password using bcrypt.
// It returns the hashed password as a string or an error if the hashing fails.
// Note: bcrypt has a maximum password length of 72 bytes, so if the input exceeds this length, it will return an error.
func HashPassword(raw string) (string, error) {
	if len(raw) > 72 {
		return "", bcrypt.ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)

	return string(bytes), err
}

// CheckPasswordHash compares a raw password with a hashed password.
// It returns true if the raw password matches the hash, false otherwise.
// If the hash is invalid or the comparison fails, it returns false.
func CheckPasswordHash(raw, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))

	return err == nil
}

// GenerateOTP generates a random one-time password of specified length.
// Default length is 6 digits if not specified.
// The OTP will be padded with leading zeros if necessary to meet the specified length.
// It returns empty string if there's an error generating random number.
func GenerateOTP(length ...int) string {
	otpLength := 6 // default length
	if len(length) > 0 && length[0] > 0 {
		otpLength = length[0]
	}

	// Calculate min and max values based on length
	min := int64(0)
	max := int64(1)
	for range otpLength {
		max *= 10
	}
	min = max / 10

	// Generate random number between min and max-1
	bg := big.NewInt(max - min)
	rng, err := rand.Int(rand.Reader, bg)
	if err != nil {
		return ""
	}

	// Add min to ensure it has the right number of digits
	otp := rng.Int64() + min

	// Convert to string and pad with zeros if needed
	otpStr := strconv.FormatInt(otp, 10)

	return otpStr
}

// GenerateULID generates a ULID (Universally Unique Lexicographically Sortable Identifier).
// If a time is provided, it uses that time to generate the ULID; otherwise, it generates a new ULID based on the current time.
// The generated ULID is returned as a lowercase string.
func GenerateULID(t *time.Time) string {
	if t != nil {
		id, err := ulid.New(ulid.Timestamp(*t), rand.Reader)

		if err == nil {
			return strings.ToLower(id.String())
		}
	}

	return strings.ToLower(ulid.Make().String())
}

// ParseULID parses a string into a ULID.
// If the string is not a valid ULID, it returns an empty ULID and an error.
// If the string is valid, it returns the parsed ULID.
func ParseULID(id string) (ulid.ULID, error) {
	parsed, err := ulid.Parse(id)
	if err != nil {
		return ulid.ULID{}, err
	}

	return parsed, nil
}

// GenerateUUID generates a new UUID (Universally Unique Identifier).
// It returns the UUID as a lowercase string.
func GenerateUUID() string {
	return strings.ToLower(uuid.New().String())
}

// ParseUUID parses a string into a UUID.
// If the string is not a valid UUID, it returns uuid.Nil and an error.
// If the string is valid, it returns the parsed UUID.
func ParseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return parsed, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// SecureCompare performs constant-time comparison of two strings
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SanitizeEmail normalizes and validates email format
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ExtractIPAddress extracts the real IP address from request
func ExtractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// IsPasswordSecure checks if password meets security requirements
func IsPasswordSecure(password string) (bool, []string) {
	var issues []string

	if len(password) < 8 {
		issues = append(issues, "Password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		issues = append(issues, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		issues = append(issues, "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		issues = append(issues, "Password must contain at least one digit")
	}
	if !hasSpecial {
		issues = append(issues, "Password must contain at least one special character")
	}

	return len(issues) == 0, issues
}
