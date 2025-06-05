package accessor

import (
	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/gin-gonic/gin"
)

// GetUserID retrieves the user ID from the http request context.
// It checks if the user ID is present in the context and returns it.
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	return userID.(string)
}

// GetDeviceInfo retrieves device information from the http request context.
// It returns the device ID, client IP, and user agent.
// If the device ID is not present in the headers, it generates a new UUID.
// The device ID is always set in the response header for consistency.
func GetDeviceInfo(c *gin.Context) (string, string, string) {
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = utils.GenerateUUID()
	}

	// Always set the device ID in the response header
	c.Header("X-Device-ID", deviceID)

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	return deviceID, ip, ua
}

// GetQueryParam retrieves a query parameter from the http request context.
// It checks if the parameter exists and returns its value.
func GetQueryParam(c *gin.Context, key string) string {
	value, exists := c.GetQuery(key)
	if !exists {
		return ""
	}

	return value
}
