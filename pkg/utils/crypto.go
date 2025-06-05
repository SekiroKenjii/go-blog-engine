package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
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
