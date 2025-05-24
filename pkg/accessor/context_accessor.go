package accessor

import (
	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	return userID.(string), true
}
