package utils

import (
	"reflect"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Test constants
const (
	testUserNamespace = "User.Username"
	testStructField   = "Struct.Field"
	testEmailValue    = "test@example.com"
)

// Mock implementation of validator.FieldError for testing
type mockFieldError struct {
	field           string
	tag             string
	param           string
	value           interface{}
	fieldType       reflect.Type
	structField     string
	structNamespace string
}

func (m mockFieldError) Tag() string                       { return m.tag }
func (m mockFieldError) ActualTag() string                 { return m.tag }
func (m mockFieldError) Namespace() string                 { return m.structNamespace }
func (m mockFieldError) StructNamespace() string           { return m.structNamespace }
func (m mockFieldError) Field() string                     { return m.field }
func (m mockFieldError) StructField() string               { return m.structField }
func (m mockFieldError) Value() interface{}                { return m.value }
func (m mockFieldError) Param() string                     { return m.param }
func (m mockFieldError) Kind() reflect.Kind                { return reflect.String }
func (m mockFieldError) Type() reflect.Type                { return m.fieldType }
func (m mockFieldError) Translate(ut ut.Translator) string { return "mock error" }
func (m mockFieldError) Error() string                     { return "mock error" }

func newMockFieldError(field, tag, param string, value interface{}, structField, structNamespace string) validator.FieldError {
	return mockFieldError{
		field:           field,
		tag:             tag,
		param:           param,
		value:           value,
		fieldType:       reflect.TypeOf(value),
		structField:     structField,
		structNamespace: structNamespace,
	}
}

func TestFormatFieldError(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		error    validator.FieldError
		expected string
	}{
		{
			name:     "replace field placeholder",
			format:   "The __FIELD__ is invalid",
			error:    newMockFieldError("username", "required", "", "", "Username", testUserNamespace),
			expected: "The username is invalid",
		},
		{
			name:     "replace tag placeholder",
			format:   "Validation failed for __TAG__ rule",
			error:    newMockFieldError("email", "email", "", "invalid-email", "Email", "User.Email"),
			expected: "Validation failed for email rule",
		},
		{
			name:     "replace param placeholder",
			format:   "Length must be at least __PARAM__ characters",
			error:    newMockFieldError("password", "min", "8", "short", "Password", "User.Password"),
			expected: "Length must be at least 8 characters",
		},
		{
			name:     "replace value placeholder",
			format:   "Value '__VALUE__' is not valid",
			error:    newMockFieldError("age", "min", "18", 15, "Age", "User.Age"),
			expected: "Value '15' is not valid",
		},
		{
			name:     "replace type placeholder",
			format:   "Expected type __TYPE__ but got something else",
			error:    newMockFieldError("count", "number", "", "not-a-number", "Count", "User.Count"),
			expected: "Expected type string but got something else",
		},
		{
			name:     "replace struct field placeholder",
			format:   "Field __STRUCT_FIELD__ has an error",
			error:    newMockFieldError("username", "required", "", "", "Username", testUserNamespace),
			expected: "Field Username has an error",
		},
		{
			name:     "replace struct name placeholder",
			format:   "Error in __STRUCT_NAME__",
			error:    newMockFieldError("username", "required", "", "", "Username", testUserNamespace),
			expected: "Error in " + testUserNamespace,
		},
		{
			name:     "multiple placeholders",
			format:   "Field __FIELD__ in __STRUCT_NAME__ failed __TAG__ validation with param __PARAM__",
			error:    newMockFieldError("email", "min", "5", "ab", "Email", "User.Email"),
			expected: "Field email in User.Email failed min validation with param 5",
		},
		{
			name:     "no placeholders",
			format:   "This is a static message",
			error:    newMockFieldError("field", "tag", "param", "value", "StructField", testStructField),
			expected: "This is a static message",
		},
		{
			name:     "unknown placeholder",
			format:   "This has __UNKNOWN__ placeholder",
			error:    newMockFieldError("field", "tag", "param", "value", "StructField", testStructField),
			expected: "This has  placeholder",
		},
		{
			name:     "nil value placeholder",
			format:   "Value is __VALUE__",
			error:    newMockFieldError("field", "required", "", nil, "Field", testStructField),
			expected: "Value is ",
		},
		{
			name:     "empty string placeholders",
			format:   "__FIELD__ __TAG__ __PARAM__ __VALUE__",
			error:    newMockFieldError("", "", "", "", "", ""),
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatFieldError(tt.format, tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPlaceholders(t *testing.T) {
	// This function is not exported, so we test it indirectly through FormatFieldError
	// But we can create a test that verifies the behavior through the public API

	tests := []struct {
		name     string
		format   string
		expected []string // Expected number of unique placeholders
	}{
		{
			name:     "single placeholder",
			format:   "Error: __FIELD__",
			expected: []string{"__FIELD__"},
		},
		{
			name:     "multiple different placeholders",
			format:   "__FIELD__ failed __TAG__ validation",
			expected: []string{"__FIELD__", "__TAG__"},
		},
		{
			name:     "duplicate placeholders",
			format:   "__FIELD__ and __FIELD__ both invalid",
			expected: []string{"__FIELD__", "__FIELD__"}, // Both instances should be found
		},
		{
			name:     "no placeholders",
			format:   "Static error message",
			expected: []string{},
		},
		{
			name:     "mixed valid and invalid placeholders",
			format:   "__FIELD__ has __INVALID_PLACEHOLDER__ error",
			expected: []string{"__FIELD__", "__INVALID_PLACEHOLDER__"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We test this indirectly by checking that known placeholders are replaced
			// and unknown ones are replaced with empty strings
			mockError := newMockFieldError("testfield", "testtag", "testparam", "testvalue", "TestField", "Test.TestField")
			result := utils.FormatFieldError(tt.format, mockError)

			// Verify that known placeholders are replaced (not empty after replacement)
			if tt.name == "single placeholder" {
				assert.Contains(t, result, "testfield")
				assert.NotContains(t, result, "__FIELD__")
			} else if tt.name == "multiple different placeholders" {
				assert.Contains(t, result, "testfield")
				assert.Contains(t, result, "testtag")
				assert.NotContains(t, result, "__FIELD__")
				assert.NotContains(t, result, "__TAG__")
			}
		})
	}
}

func TestToString(t *testing.T) {
	// This function is not exported, but we can test it indirectly through FormatFieldError
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "string value",
			value:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "integer value",
			value:    42,
			expected: "42",
		},
		{
			name:     "float value",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean true",
			value:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			value:    false,
			expected: "false",
		},
		{
			name:     "nil value",
			value:    nil,
			expected: "",
		},
		{
			name:     "string with spaces",
			value:    "  Spaced String  ",
			expected: "spaced string",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test toString indirectly through FormatFieldError with __VALUE__ placeholder
			mockError := newMockFieldError("field", "tag", "param", tt.value, "Field", testStructField)
			result := utils.FormatFieldError("__VALUE__", mockError)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildPlaceholderMap(t *testing.T) {
	// Test this indirectly by verifying behavior through FormatFieldError

	t.Run("all placeholders mapped correctly", func(t *testing.T) {
		mockError := newMockFieldError("username", "required", "true", testEmailValue, "Username", testUserNamespace)

		// Test each placeholder individually
		testCases := []struct {
			placeholder string
			expected    string
		}{
			{"__FIELD__", "username"},
			{"__TAG__", "required"},
			{"__PARAM__", "true"},
			{"__VALUE__", testEmailValue},
			{"__TYPE__", "string"},
			{"__STRUCT_FIELD__", "Username"},
			{"__STRUCT_NAME__", testUserNamespace},
		}

		for _, tc := range testCases {
			result := utils.FormatFieldError(tc.placeholder, mockError)
			assert.Equal(t, tc.expected, result, "Failed for placeholder %s", tc.placeholder)
		}
	})

	t.Run("unknown placeholder returns empty", func(t *testing.T) {
		mockError := newMockFieldError("field", "tag", "param", "value", "Field", testStructField)
		result := utils.FormatFieldError("__UNKNOWN__", mockError)
		assert.Equal(t, "", result)
	})
}

// Integration test with real validator
func TestFormatFieldErrorWithRealValidator(t *testing.T) {
	type User struct {
		Name  string `validate:"required,min=3"`
		Email string `validate:"required,email"`
		Age   int    `validate:"min=18"`
	}

	validate := validator.New()

	user := User{
		Name:  "Jo",            // Too short
		Email: "invalid-email", // Invalid format
		Age:   15,              // Too young
	}

	err := validate.Struct(user)
	assert.Error(t, err)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			// Test with different format strings
			result1 := utils.FormatFieldError("Field __FIELD__ failed __TAG__ validation", fieldError)
			assert.Contains(t, result1, fieldError.Field())
			assert.Contains(t, result1, fieldError.Tag())

			result2 := utils.FormatFieldError("__VALUE__ is not valid for __STRUCT_FIELD__", fieldError)
			assert.Contains(t, result2, fieldError.StructField())
		}
	}
}

// Benchmark tests
func BenchmarkFormatFieldError(b *testing.B) {
	format := "Field __FIELD__ in __STRUCT_NAME__ failed __TAG__ validation with value __VALUE__"
	mockError := newMockFieldError("username", "required", "true", testEmailValue, "Username", testUserNamespace)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.FormatFieldError(format, mockError)
	}
}

func BenchmarkFormatFieldErrorSimple(b *testing.B) {
	format := "__FIELD__ is required"
	mockError := newMockFieldError("username", "required", "", "", "Username", testUserNamespace)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.FormatFieldError(format, mockError)
	}
}

func BenchmarkFormatFieldErrorNoPlaceholders(b *testing.B) {
	format := "This is a static error message"
	mockError := newMockFieldError("username", "required", "", "", "Username", testUserNamespace)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.FormatFieldError(format, mockError)
	}
}
