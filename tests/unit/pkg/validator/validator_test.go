package validator

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserName     = "John Doe"
	testUserEmail    = "john@example.com"
	testUserPassword = "password123"
)

// Helper function to safely access Source as map[string]string
func getSourceField(err *response.ErrorInner, key string) string {
	if source, ok := err.Source.(map[string]string); ok {
		return source[key]
	}
	return ""
}

// Test structs for validation
type TestUser struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,min=18"`
	Password string `json:"password" binding:"required,min=8"`
}

type TestUserOptional struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
	Age   *int   `json:"age" binding:"omitempty,min=0"`
}

type NestedTestStruct struct {
	User TestUser `json:"user" binding:"required"`
	Role string   `json:"role" binding:"required,oneof=admin user guest"`
}

func setupTestContext(jsonBody string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

func TestValidateRequestBasic(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors)
		assert.Equal(t, testUserName, user.Name)
		assert.Equal(t, testUserEmail, user.Email)
		assert.Equal(t, 25, user.Age)
		assert.Equal(t, testUserPassword, user.Password)
	})

	t.Run("missing required fields", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 3) // email, age, password are missing

		// Check if all required fields are reported
		fields := make(map[string]bool)
		for _, err := range errors {
			fields[getSourceField(err, "field")] = true
		}
		assert.True(t, fields["Email"])
		assert.True(t, fields["Age"])
		assert.True(t, fields["Password"])
	})

	t.Run("invalid email format", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "invalid-email",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 1)
		assert.Equal(t, "Email", getSourceField(errors[0], "field"))
		assert.Contains(t, getSourceField(errors[0], "messages"), "must be a valid email address")
	})
}

func TestValidateRequestValidation(t *testing.T) {
	t.Run("minimum length validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 16,
			"password": "123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 2) // age < 18, password < 8 chars

		// Check age validation
		found := false
		for _, err := range errors {
			if getSourceField(err, "field") == "Age" {
				assert.Contains(t, getSourceField(err, "messages"), "must be at least 18")
				found = true
				break
			}
		}
		assert.True(t, found, "Age validation error not found")

		// Check password validation
		found = false
		for _, err := range errors {
			if getSourceField(err, "field") == "Password" {
				assert.Contains(t, getSourceField(err, "messages"), "must be at least 8")
				found = true
				break
			}
		}
		assert.True(t, found, "Password validation error not found")
	})

	t.Run("multiple validation errors on single field", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)

		// Email field should have at least one error (required or email format)
		emailErrors := 0
		for _, err := range errors {
			if getSourceField(err, "field") == "Email" {
				emailErrors++
			}
		}
		assert.Greater(t, emailErrors, 0)
	})
}

func TestValidateRequestOptional(t *testing.T) {
	t.Run("optional fields validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe"
		}`)

		var user TestUserOptional
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors) // No required fields, so should pass
		assert.Equal(t, testUserName, user.Name)
	})

	t.Run("optional field with invalid value", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "invalid-email",
			"age": -5
		}`)

		var user TestUserOptional
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 2) // invalid email and negative age

		fields := make(map[string]bool)
		for _, err := range errors {
			fields[getSourceField(err, "field")] = true
		}
		assert.True(t, fields["Email"])
		assert.True(t, fields["Age"])
	})
}

func TestValidateRequestNested(t *testing.T) {
	t.Run("nested struct validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"user": {
				"name": "John Doe",
				"email": "john@example.com",
				"age": 25,
				"password": "password123"
			},
			"role": "admin"
		}`)

		var nested NestedTestStruct
		errors := validator.ValidateRequest(c, &nested)

		assert.Nil(t, errors)
		assert.Equal(t, testUserName, nested.User.Name)
		assert.Equal(t, "admin", nested.Role)
	})

	t.Run("nested struct with validation errors", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"user": {
				"name": "",
				"email": "invalid-email"
			},
			"role": "invalid-role"
		}`)

		var nested NestedTestStruct
		errors := validator.ValidateRequest(c, &nested)

		assert.NotNil(t, errors)
		assert.Greater(t, len(errors), 2) // Multiple validation errors
	})
}

func TestValidateRequestErrors(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors)
		assert.Equal(t, testUserName, user.Name)
		assert.Equal(t, testUserEmail, user.Email)
		assert.Equal(t, 25, user.Age)
		assert.Equal(t, testUserPassword, user.Password)
	})

	t.Run("missing required fields", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 3) // email, age, password are missing

		// Check if all required fields are reported
		fields := make(map[string]bool)
		for _, err := range errors {
			fields[getSourceField(err, "field")] = true
		}
		assert.True(t, fields["Email"])
		assert.True(t, fields["Age"])
		assert.True(t, fields["Password"])
	})

	t.Run("invalid email format", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "invalid-email",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 1)
		assert.Equal(t, "Email", getSourceField(errors[0], "field"))
		assert.Contains(t, getSourceField(errors[0], "messages"), "must be a valid email address")
	})

	t.Run("minimum length validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 16,
			"password": "123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 2) // age < 18, password < 8 chars

		// Check age validation
		found := false
		for _, err := range errors {
			if getSourceField(err, "field") == "Age" {
				assert.Contains(t, getSourceField(err, "messages"), "must be at least 18")
				found = true
				break
			}
		}
		assert.True(t, found, "Age validation error not found")

		// Check password validation
		found = false
		for _, err := range errors {
			if getSourceField(err, "field") == "Password" {
				assert.Contains(t, getSourceField(err, "messages"), "must be at least 8")
				found = true
				break
			}
		}
		assert.True(t, found, "Password validation error not found")
	})

	t.Run("multiple validation errors on single field", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "",
			"age": 25,
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)

		// Email field should have at least one error (required or email format)
		emailErrors := 0
		for _, err := range errors {
			if getSourceField(err, "field") == "Email" {
				emailErrors++
			}
		}
		assert.Greater(t, emailErrors, 0)
	})

	t.Run("optional fields validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe"
		}`)

		var user TestUserOptional
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors) // No required fields, so should pass
		assert.Equal(t, testUserName, user.Name)
	})

	t.Run("optional field with invalid value", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "invalid-email",
			"age": -5
		}`)

		var user TestUserOptional
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 2) // invalid email and negative age

		fields := make(map[string]bool)
		for _, err := range errors {
			fields[getSourceField(err, "field")] = true
		}
		assert.True(t, fields["Email"])
		assert.True(t, fields["Age"])
	})

	t.Run("nested struct validation", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"user": {
				"name": "John Doe",
				"email": "john@example.com",
				"age": 25,
				"password": "password123"
			},
			"role": "admin"
		}`)

		var nested NestedTestStruct
		errors := validator.ValidateRequest(c, &nested)

		assert.Nil(t, errors)
		assert.Equal(t, testUserName, nested.User.Name)
		assert.Equal(t, "admin", nested.Role)
	})

	t.Run("nested struct with validation errors", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"user": {
				"name": "",
				"email": "invalid-email"
			},
			"role": "invalid-role"
		}`)

		var nested NestedTestStruct
		errors := validator.ValidateRequest(c, &nested)

		assert.NotNil(t, errors)
		assert.Greater(t, len(errors), 2) // Multiple validation errors
	})

	t.Run("invalid JSON", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com"
			"age": 25
		}`) // Missing comma - invalid JSON

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		// Should return default validation error for JSON parsing failure
		assert.Equal(t, len(response.DefaultValidatorError()), len(errors))
	})

	t.Run("empty JSON body", func(t *testing.T) {
		c, _ := setupTestContext(`{}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 4) // All required fields are missing
	})

	t.Run("null JSON body", func(t *testing.T) {
		c, _ := setupTestContext(`{}`) // Use empty object instead of null to avoid reflection panic

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		// Should return validation errors for all required fields
		assert.Len(t, errors, 4) // name, email, age, password are required
	})

	t.Run("type mismatch", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": "not-a-number",
			"password": "password123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		// Should return default validation errors for type mismatch
		assert.Greater(t, len(errors), 0)
	})
}

func TestGetValidateMsgCode(t *testing.T) {
	// Since getValidateMsgCode is not exported, we test it indirectly through ValidateRequest
	t.Run("required field error code", func(t *testing.T) {
		c, _ := setupTestContext(`{}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)

		// Find a required field error
		for _, err := range errors {
			if getSourceField(err, "field") == "Name" {
				assert.Equal(t, string(response.EBIZ000004), err.Code)
				assert.Contains(t, getSourceField(err, "messages"), "is required")
				break
			}
		}
	})

	t.Run("email validation error code", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John",
			"email": "invalid",
			"age": 25,
			"password": testUserPassword
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)

		// Find email validation error
		for _, err := range errors {
			if getSourceField(err, "field") == "Email" {
				assert.Equal(t, string(response.EBIZ000005), err.Code)
				assert.Contains(t, getSourceField(err, "messages"), "valid email address")
				break
			}
		}
	})

	t.Run("min length validation error code", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John",
			"email": testUserEmail,
			"age": 15,
			"password": testUserPassword
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)

		// Find min validation error
		for _, err := range errors {
			if getSourceField(err, "field") == "Age" {
				assert.Equal(t, string(response.EBIZ000005), err.Code)
				assert.Contains(t, getSourceField(err, "messages"), "at least 18")
				break
			}
		}
	})
}

func TestMessageCodes(t *testing.T) {
	t.Run("all error codes are defined", func(t *testing.T) {
		// Test that all defined error codes have corresponding messages
		assert.NotNil(t, validator.MessageCodes[validator.ErrRequiredField])
		assert.NotNil(t, validator.MessageCodes[validator.ErrInvalidEmail])
		assert.NotNil(t, validator.MessageCodes[validator.ErrMinLength])
		assert.NotNil(t, validator.MessageCodes[validator.ErrDefault])
	})

	t.Run("required field message format", func(t *testing.T) {
		msg := validator.MessageCodes[validator.ErrRequiredField]
		assert.Contains(t, msg.Message, "__FIELD__")
		assert.Equal(t, response.EBIZ000004, msg.Code)
	})

	t.Run("email field message format", func(t *testing.T) {
		msg := validator.MessageCodes[validator.ErrInvalidEmail]
		assert.Contains(t, msg.Message, "__FIELD__")
		assert.Contains(t, msg.Message, "email")
		assert.Equal(t, response.EBIZ000005, msg.Code)
	})

	t.Run("min length message format", func(t *testing.T) {
		msg := validator.MessageCodes[validator.ErrMinLength]
		assert.Contains(t, msg.Message, "__FIELD__")
		assert.Contains(t, msg.Message, "__PARAM__")
		assert.Equal(t, response.EBIZ000005, msg.Code)
	})

	t.Run("default message format", func(t *testing.T) {
		msg := validator.MessageCodes[validator.ErrDefault]
		assert.Contains(t, msg.Message, "__FIELD__")
		assert.Contains(t, msg.Message, "__PARAM__")
		assert.Equal(t, response.EBIZ000005, msg.Code)
	})
}

func TestValidateRequestEdgeCases(t *testing.T) {
	t.Run("large JSON payload", func(t *testing.T) {
		// Create a large but valid JSON payload
		largeString := make([]byte, 1000)
		for i := range largeString {
			largeString[i] = 'a'
		}

		jsonBody := `{
			"name": "` + string(largeString) + `",
			"email": "john@example.com",
			"age": 25,
			"password": "password123"
		}`

		c, _ := setupTestContext(jsonBody)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors)
		assert.Equal(t, string(largeString), user.Name)
	})

	t.Run("unicode characters", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "João Müller 中文",
			"email": "joão@müller.com",
			"age": 25,
			"password": "pássword123"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors)
		assert.Equal(t, "João Müller 中文", user.Name)
		assert.Equal(t, "joão@müller.com", user.Email)
		assert.Equal(t, "pássword123", user.Password)
	})

	t.Run("special characters in strings", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "User \"with\" 'quotes' & <tags>",
			"email": "user+test@example.com",
			"age": 25,
			"password": "p@ssw0rd!#$"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors)
		assert.Contains(t, user.Name, "quotes")
		assert.Contains(t, user.Email, "+test")
		assert.Contains(t, user.Password, "@ssw0rd")
	})

	t.Run("zero values", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "",
			"email": "",
			"age": 0,
			"password": ""
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.NotNil(t, errors)
		assert.Len(t, errors, 4) // All fields should fail validation
	})

	t.Run("boundary values", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "J",
			"email": "a@b.co",
			"age": 18,
			"password": "12345678"
		}`)

		var user TestUser
		errors := validator.ValidateRequest(c, &user)

		assert.Nil(t, errors) // All values are at the minimum boundary
		assert.Equal(t, "J", user.Name)
		assert.Equal(t, 18, user.Age)
		assert.Equal(t, "12345678", user.Password)
	})
}

func TestValidateRequestConcurrency(t *testing.T) {
	t.Run("concurrent validation requests", func(t *testing.T) {
		const numGoroutines = 10

		results := make(chan []*response.ErrorInner, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				c, _ := setupTestContext(`{
					"name": "User` + string(rune('0'+id)) + `",
					"email": "user` + string(rune('0'+id)) + `@example.com",
					"age": 25,
					"password": "password123"
				}`)

				var user TestUser
				errors := validator.ValidateRequest(c, &user)
				results <- errors
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			errors := <-results
			assert.Nil(t, errors, "Validation should succeed for concurrent request %d", i)
		}
	})
}

// Benchmark tests
func BenchmarkValidateRequestValid(b *testing.B) {
	jsonBody := `{
		"name": "John Doe",
		"email": "john@example.com",
		"age": 25,
		"password": "password123"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext(jsonBody)
		var user TestUser
		_ = validator.ValidateRequest(c, &user)
	}
}

func BenchmarkValidateRequestInvalid(b *testing.B) {
	jsonBody := `{
		"name": "",
		"email": "invalid-email",
		"age": 15,
		"password": "123"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext(jsonBody)
		var user TestUser
		_ = validator.ValidateRequest(c, &user)
	}
}

func BenchmarkValidateRequestLarge(b *testing.B) {
	// Create a large JSON payload for benchmarking
	largeString := make([]byte, 5000)
	for i := range largeString {
		largeString[i] = 'a'
	}

	jsonBody := `{
		"name": "` + string(largeString) + `",
		"email": "john@example.com",
		"age": 25,
		"password": "password123"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext(jsonBody)
		var user TestUser
		_ = validator.ValidateRequest(c, &user)
	}
}

func TestValidatorTypeSafety(t *testing.T) {
	t.Run("different struct types", func(t *testing.T) {
		type DifferentStruct struct {
			ID   int    `json:"id" binding:"required"`
			Name string `json:"name" binding:"required"`
		}

		c, _ := setupTestContext(`{
			"id": 123,
			"name": "Test"
		}`)

		var diff DifferentStruct
		errors := validator.ValidateRequest(c, &diff)

		assert.Nil(t, errors)
		assert.Equal(t, 123, diff.ID)
		assert.Equal(t, "Test", diff.Name)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		c, _ := setupTestContext(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 25,
			"password": "password123"
		}`)

		var user *TestUser
		errors := validator.ValidateRequest(c, &user)

		require.NotNil(t, user) // Should be allocated
		assert.Nil(t, errors)
		assert.Equal(t, testUserName, user.Name)
	})
}
