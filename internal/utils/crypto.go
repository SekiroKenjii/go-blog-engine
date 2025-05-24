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

func HashSHA256(raw string) string {
	h := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(h[:])
}

func HashPassword(raw string) (string, error) {
	if len(raw) > 72 {
		return "", bcrypt.ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)

	return string(bytes), err
}

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

func GenerateULID(t *time.Time) string {
	if t != nil {
		id, err := ulid.New(ulid.Timestamp(*t), rand.Reader)

		if err == nil {
			return strings.ToLower(id.String())
		}
	}

	return strings.ToLower(ulid.Make().String())
}

func GenerateUUID() string {
	return strings.ToLower(uuid.New().String())
}
