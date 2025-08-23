package abstract

import "github.com/gin-gonic/gin"

type IHandler interface {
	// RegisterPublicRoutes registers routes that don't require authentication.
	// These routes are accessible to all users without authentication.
	RegisterPublicRoutes(*gin.RouterGroup)

	// RegisterProtectedRoutes registers routes that require authentication.
	// These routes are only accessible to authenticated users.
	RegisterProtectedRoutes(*gin.RouterGroup)
}
