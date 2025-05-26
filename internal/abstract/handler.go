package abstract

import "github.com/gin-gonic/gin"

type IHandler interface {
	// RegisterRoutes registers the routes for the handler.
	RegisterRoutes(*gin.RouterGroup)
}
