package accessor

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/accessor"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	testUserAgent     = "test-agent"
	deviceIDHeader    = "X-Device-ID"
	userAgentHeader   = "User-Agent"
	integrationDevice = "integration-device"
)

func setupGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func(*gin.Context)
		expected string
	}{
		{
			name: "returns user ID when set in context",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "user123")
			},
			expected: "user123",
		},
		{
			name: "returns empty string when user ID not set",
			setupCtx: func(c *gin.Context) {
				// Don't set anything
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupGinContext()
			tt.setupCtx(c)

			result := accessor.GetUserID(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDeviceInfo(t *testing.T) {
	tests := []struct {
		name           string
		setupCtx       func(*gin.Context)
		expectedID     string
		expectedIP     string
		expectedUA     string
		checkGenerated bool
	}{
		{
			name: "returns device info when X-Device-ID header is set",
			setupCtx: func(c *gin.Context) {
				c.Request.Header.Set(deviceIDHeader, "device123")
				c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")
				c.Request.Header.Set(userAgentHeader, testUserAgent)
			},
			expectedID: "device123",
			expectedIP: "192.168.1.1",
			expectedUA: testUserAgent,
		},
		{
			name: "generates new UUID when X-Device-ID header not set",
			setupCtx: func(c *gin.Context) {
				// No device ID header set - should generate new UUID
			},
			expectedIP:     "192.0.2.1", // Default test IP from gin
			expectedUA:     "",
			checkGenerated: true,
		},
		{
			name: "handles missing headers gracefully",
			setupCtx: func(c *gin.Context) {
				// No headers set
			},
			expectedIP:     "192.0.2.1", // Default test IP from gin
			expectedUA:     "",
			checkGenerated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupGinContext()
			c.Request = httptest.NewRequest("GET", "/", nil)
			tt.setupCtx(c)

			deviceID, clientIP, userAgent := accessor.GetDeviceInfo(c)

			if tt.checkGenerated {
				assert.NotEmpty(t, deviceID)
				// Validate that deviceID is a valid UUID
				_, err := uuid.Parse(deviceID)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedID, deviceID)
			}

			assert.Equal(t, tt.expectedIP, clientIP)
			assert.Equal(t, tt.expectedUA, userAgent)

			// Check that response header is set
			assert.Equal(t, deviceID, c.Writer.Header().Get(deviceIDHeader))
		})
	}
}

func TestGetQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		key         string
		expected    string
	}{
		{
			name:        "returns query parameter when exists",
			queryParams: "param1=value1&param2=value2",
			key:         "param1",
			expected:    "value1",
		},
		{
			name:        "returns empty string when parameter not exists",
			queryParams: "param1=value1&param2=value2",
			key:         "param3",
			expected:    "",
		},
		{
			name:        "returns empty string when parameter exists but empty",
			queryParams: "param1=&param2=value2",
			key:         "param1",
			expected:    "",
		},
		{
			name:        "returns empty string when no query parameters",
			queryParams: "",
			key:         "param1",
			expected:    "",
		},
		{
			name:        "returns parameter with special characters",
			queryParams: "search=hello%20world&type=test",
			key:         "search",
			expected:    "hello world",
		},
		{
			name:        "returns first value when parameter appears multiple times",
			queryParams: "param1=first&param1=second",
			key:         "param1",
			expected:    "first",
		},
		{
			name:        "handles case sensitive parameter names",
			queryParams: "Param1=value1&param1=value2",
			key:         "param1",
			expected:    "value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupGinContext()

			// Set up request with query parameters
			req := httptest.NewRequest("GET", "/?"+tt.queryParams, nil)
			c.Request = req

			result := accessor.GetQueryParam(c, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetQueryParamWithComplexScenarios(t *testing.T) {
	t.Run("handles URL encoded values correctly", func(t *testing.T) {
		c, _ := setupGinContext()

		// Test with URL encoded values
		queryValues := url.Values{}
		queryValues.Add("email", "user@example.com")
		queryValues.Add("message", "Hello World!")
		queryValues.Add("special", "value with spaces & symbols")

		req := httptest.NewRequest("GET", "/?"+queryValues.Encode(), nil)
		c.Request = req

		assert.Equal(t, "user@example.com", accessor.GetQueryParam(c, "email"))
		assert.Equal(t, "Hello World!", accessor.GetQueryParam(c, "message"))
		assert.Equal(t, "value with spaces & symbols", accessor.GetQueryParam(c, "special"))
	})

	t.Run("returns empty string for missing parameter", func(t *testing.T) {
		c, _ := setupGinContext()
		req := httptest.NewRequest("GET", "/", nil)
		c.Request = req

		result := accessor.GetQueryParam(c, "nonexistent")
		assert.Equal(t, "", result)
	})

	t.Run("handles empty query string gracefully", func(t *testing.T) {
		c, _ := setupGinContext()
		req := httptest.NewRequest("GET", "/?", nil)
		c.Request = req

		result := accessor.GetQueryParam(c, "param")
		assert.Equal(t, "", result)
	})
}

func TestGetUserIDIntegration(t *testing.T) {
	// Test integration with actual Gin context flow
	t.Run("user ID set by middleware", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		var capturedUserID string

		router.Use(func(c *gin.Context) {
			c.Set("user_id", "middleware-user-123")
			c.Next()
		})

		router.GET("/test", func(c *gin.Context) {
			capturedUserID = accessor.GetUserID(c)
			c.JSON(200, gin.H{"success": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, "middleware-user-123", capturedUserID)
		assert.Equal(t, 200, w.Code)
	})
}

func TestGetDeviceInfoIntegration(t *testing.T) {
	// Test integration with actual Gin context flow
	t.Run("device info set by headers", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		var capturedDeviceID, capturedIP, capturedUA string

		router.GET("/test", func(c *gin.Context) {
			capturedDeviceID, capturedIP, capturedUA = accessor.GetDeviceInfo(c)
			c.JSON(200, gin.H{"success": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(deviceIDHeader, integrationDevice)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		req.Header.Set(userAgentHeader, testUserAgent)
		router.ServeHTTP(w, req)

		assert.Equal(t, integrationDevice, capturedDeviceID)
		assert.Equal(t, "10.0.0.1", capturedIP)
		assert.Equal(t, testUserAgent, capturedUA)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, integrationDevice, w.Header().Get(deviceIDHeader))
	})
}

func TestGetQueryParamEdgeCases(t *testing.T) {
	t.Run("handles malformed query parameters", func(t *testing.T) {
		c, _ := setupGinContext()
		req := httptest.NewRequest("GET", "/?param1=value1&malformed&param2=value2", nil)
		c.Request = req

		// Should still work for well-formed parameters
		assert.Equal(t, "value1", accessor.GetQueryParam(c, "param1"))
		assert.Equal(t, "value2", accessor.GetQueryParam(c, "param2"))
		assert.Equal(t, "", accessor.GetQueryParam(c, "malformed"))
	})

	t.Run("handles parameters with equals in value", func(t *testing.T) {
		c, _ := setupGinContext()
		req := httptest.NewRequest("GET", "/?equation=a=b&normal=test", nil)
		c.Request = req

		assert.Equal(t, "a=b", accessor.GetQueryParam(c, "equation"))
		assert.Equal(t, "test", accessor.GetQueryParam(c, "normal"))
	})
}

// Benchmark tests
func BenchmarkGetUserID(b *testing.B) {
	c, _ := setupGinContext()
	c.Set("user_id", "benchmark-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.GetUserID(c)
	}
}

func BenchmarkGetDeviceInfo(b *testing.B) {
	c, _ := setupGinContext()
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set(deviceIDHeader, "benchmark-device")
	c.Request.Header.Set(userAgentHeader, testUserAgent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.GetDeviceInfo(c)
	}
}

func BenchmarkGetQueryParam(b *testing.B) {
	c, _ := setupGinContext()
	req := httptest.NewRequest("GET", "/?param1=value1&param2=value2", nil)
	c.Request = req

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.GetQueryParam(c, "param1")
	}
}
