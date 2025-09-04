package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testData = "test data"
)

func setupGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		message      string
		data         interface{}
		warnings     []*any
		expectedCode int
	}{
		{
			name:         "success with string data",
			statusCode:   http.StatusOK,
			message:      "Success message",
			data:         testData,
			warnings:     nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "success with map data",
			statusCode:   http.StatusCreated,
			message:      "Created successfully",
			data:         map[string]string{"id": "123", "name": "test"},
			warnings:     nil,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "success with warnings",
			statusCode:   http.StatusOK,
			message:      "Success with warnings",
			data:         "data",
			warnings:     []*any{},
			expectedCode: http.StatusOK,
		},
		{
			name:         "success with nil data",
			statusCode:   http.StatusNoContent,
			message:      "No content",
			data:         (*string)(nil),
			warnings:     nil,
			expectedCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext()

			response.Success(c, tt.statusCode, tt.message, &tt.data, tt.warnings)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.message)

			if tt.data != nil {
				assert.Contains(t, w.Body.String(), `"data":`)
			}

			assert.Contains(t, w.Body.String(), `"errors":null`)
		})
	}
}

func TestFailure(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		errorCode    response.ErrorCode
		errors       []*response.ErrorInner
		warnings     []*any
		expectedCode int
	}{
		{
			name:       "simple failure",
			statusCode: http.StatusBadRequest,
			errorCode:  response.EBIZ000001,
			errors: []*response.ErrorInner{
				{Code: "TEST001", Source: map[string]string{"field": "test"}},
			},
			warnings:     nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:       "failure with multiple errors",
			statusCode: http.StatusUnprocessableEntity,
			errorCode:  response.EBIZ000002,
			errors: []*response.ErrorInner{
				{Code: "TEST001", Source: map[string]string{"field": "field1"}},
				{Code: "TEST002", Source: map[string]string{"field": "field2"}},
			},
			warnings:     nil,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "failure with warnings",
			statusCode:   http.StatusBadRequest,
			errorCode:    response.EBIZ000001,
			errors:       nil,
			warnings:     []*any{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext()

			response.Failure(c, tt.statusCode, tt.errorCode, tt.errors, tt.warnings)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), `"data":null`)
			assert.Contains(t, w.Body.String(), `"errors":`)

			if len(tt.errors) > 0 {
				for _, err := range tt.errors {
					assert.Contains(t, w.Body.String(), err.Code)
				}
			}
		})
	}
}

func TestNotImplemented(t *testing.T) {
	c, w := setupGinContext()

	response.NotImplemented(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Contains(t, w.Body.String(), `"data":null`)
	assert.Contains(t, w.Body.String(), string(response.FATA000002))
}

func TestTooManyRequest(t *testing.T) {
	c, w := setupGinContext()
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:8080"

	response.TooManyRequest(c)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), string(response.ESYS000010))
	assert.Contains(t, w.Body.String(), "/test")
}

func TestAuthorizationHeaderError(t *testing.T) {
	c, w := setupGinContext()

	response.AuthorizationHeaderError(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), string(response.ESYS000011))
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name              string
		defaultErrorCodes []response.ErrorCode
		expectedCode      response.ErrorCode
	}{
		{
			name:         "forbidden with default error code",
			expectedCode: response.EBIZ000003,
		},
		{
			name:              "forbidden with custom error code",
			defaultErrorCodes: []response.ErrorCode{response.EBIZ000001},
			expectedCode:      response.EBIZ000001,
		},
		{
			name:              "forbidden with multiple error codes uses first",
			defaultErrorCodes: []response.ErrorCode{response.EBIZ000002, response.EBIZ000001},
			expectedCode:      response.EBIZ000002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext()

			response.Forbidden(c, tt.defaultErrorCodes...)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), string(tt.expectedCode))
		})
	}
}

func TestHandleBizFailure(t *testing.T) {
	tests := []struct {
		name               string
		errorCode          response.ErrorCode
		defaultStatusCodes []int
		expectedStatus     int
	}{
		{
			name:           "business error with default status",
			errorCode:      response.EBIZ000001,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:               "business error with custom status",
			errorCode:          response.EBIZ000002,
			defaultStatusCodes: []int{http.StatusUnprocessableEntity},
			expectedStatus:     http.StatusUnprocessableEntity,
		},
		{
			name:           "fatal error always returns 500",
			errorCode:      response.FATA000001,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:               "fatal error ignores custom status",
			errorCode:          response.FATA000002,
			defaultStatusCodes: []int{http.StatusBadRequest},
			expectedStatus:     http.StatusInternalServerError,
		},
		{
			name:               "multiple status codes uses first",
			errorCode:          response.EBIZ000001,
			defaultStatusCodes: []int{http.StatusConflict, http.StatusUnprocessableEntity},
			expectedStatus:     http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext()

			response.HandleBizFailure(c, tt.errorCode, tt.defaultStatusCodes...)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), string(tt.errorCode))
			assert.Contains(t, w.Body.String(), `"data":null`)
		})
	}
}

func TestDefaultValidatorError(t *testing.T) {
	errors := response.DefaultValidatorError()

	require.NotNil(t, errors)
	require.Len(t, errors, 1)

	err := errors[0]
	assert.Empty(t, err.Code)

	source, ok := err.Source.(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "", source["field"])
	assert.Equal(t, "Invalid request", source["messages"])
}

func TestResponseStructures(t *testing.T) {
	t.Run("Response structure with generic data", func(t *testing.T) {
		resp := response.Response[string]{
			Message:  "test message",
			Data:     testData,
			Warnings: nil,
			Errors:   nil,
		}

		assert.Equal(t, "test message", resp.Message)
		assert.Equal(t, testData, resp.Data)
		assert.Nil(t, resp.Warnings)
		assert.Nil(t, resp.Errors)
	})

	t.Run("ErrorInner structure", func(t *testing.T) {
		errInner := &response.ErrorInner{
			Code:   "TEST001",
			Source: map[string]string{"field": "test"},
		}

		assert.Equal(t, "TEST001", errInner.Code)

		source, ok := errInner.Source.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "test", source["field"])
	})
}

func TestErrorCodeConstants(t *testing.T) {
	// Test that error codes have expected string values
	assert.Equal(t, "SBIZ000001", string(response.SBIZ000001))
	assert.Equal(t, "EBIZ000001", string(response.EBIZ000001))
	assert.Equal(t, "EBIZ000002", string(response.EBIZ000002))
	assert.Equal(t, "EBIZ000003", string(response.EBIZ000003))
	assert.Equal(t, "ESYS000010", string(response.ESYS000010))
	assert.Equal(t, "ESYS000011", string(response.ESYS000011))
	assert.Equal(t, "FATA000001", string(response.FATA000001))
	assert.Equal(t, "FATA000002", string(response.FATA000002))
}

func TestGinContextAbort(t *testing.T) {
	// Test that Failure and related functions properly abort the context
	t.Run("Failure aborts context", func(t *testing.T) {
		c, _ := setupGinContext()

		response.Failure(c, http.StatusBadRequest, response.EBIZ000001, nil, nil)

		assert.True(t, c.IsAborted())
	})

	t.Run("NotImplemented aborts context", func(t *testing.T) {
		c, _ := setupGinContext()

		response.NotImplemented(c)

		assert.True(t, c.IsAborted())
	})

	t.Run("TooManyRequest aborts context", func(t *testing.T) {
		c, _ := setupGinContext()
		c.Request = httptest.NewRequest("GET", "/test", nil)

		response.TooManyRequest(c)

		assert.True(t, c.IsAborted())
	})

	t.Run("AuthorizationHeaderError aborts context", func(t *testing.T) {
		c, _ := setupGinContext()

		response.AuthorizationHeaderError(c)

		assert.True(t, c.IsAborted())
	})

	t.Run("Forbidden aborts context", func(t *testing.T) {
		c, _ := setupGinContext()

		response.Forbidden(c)

		assert.True(t, c.IsAborted())
	})

	t.Run("HandleBizFailure aborts context", func(t *testing.T) {
		c, _ := setupGinContext()

		response.HandleBizFailure(c, response.EBIZ000001)

		assert.True(t, c.IsAborted())
	})

	t.Run("Success does not abort context", func(t *testing.T) {
		c, _ := setupGinContext()
		data := "test"

		response.Success(c, http.StatusOK, "success", &data, nil)

		assert.False(t, c.IsAborted())
	})
}

func TestComplexScenarios(t *testing.T) {
	t.Run("success with complex data structure", func(t *testing.T) {
		c, w := setupGinContext()

		complexData := map[string]interface{}{
			"users": []map[string]string{
				{"id": "1", "name": "John"},
				{"id": "2", "name": "Jane"},
			},
			"meta": map[string]int{
				"total": 2,
				"page":  1,
			},
		}

		response.Success(c, http.StatusOK, "Users retrieved", &complexData, nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Users retrieved")
		assert.Contains(t, w.Body.String(), "John")
		assert.Contains(t, w.Body.String(), "Jane")
		assert.Contains(t, w.Body.String(), `"total":2`)
	})

	t.Run("failure with complex error structure", func(t *testing.T) {
		c, w := setupGinContext()

		errors := []*response.ErrorInner{
			{
				Code: "VALIDATION_001",
				Source: map[string]string{
					"field":   "email",
					"message": "Invalid email format",
					"value":   "invalid-email",
				},
			},
			{
				Code: "VALIDATION_002",
				Source: map[string]string{
					"field":   "password",
					"message": "Password too short",
					"min":     "8",
				},
			},
		}

		response.Failure(c, http.StatusUnprocessableEntity, response.EBIZ000002, errors, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, w.Body.String(), "VALIDATION_001")
		assert.Contains(t, w.Body.String(), "VALIDATION_002")
		assert.Contains(t, w.Body.String(), "Invalid email format")
		assert.Contains(t, w.Body.String(), "Password too short")
	})
}

// Benchmark tests
func BenchmarkSuccess(b *testing.B) {
	gin.SetMode(gin.TestMode)
	data := "benchmark data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupGinContext()
		response.Success(c, http.StatusOK, "success", &data, nil)
	}
}

func BenchmarkFailure(b *testing.B) {
	gin.SetMode(gin.TestMode)
	errors := []*response.ErrorInner{
		{Code: "TEST", Source: map[string]string{"field": "test"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupGinContext()
		response.Failure(c, http.StatusBadRequest, response.EBIZ000001, errors, nil)
	}
}

func BenchmarkHandleBizFailure(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupGinContext()
		response.HandleBizFailure(c, response.EBIZ000001)
	}
}
