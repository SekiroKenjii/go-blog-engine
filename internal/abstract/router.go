package abstract

import "github.com/gin-gonic/gin"

type IRouter interface {
	// Engine returns the gin engine instance.
	Engine() *gin.Engine

	// Configure sets up the routes, middlewares, and other configurations for the router.
	// It should be called after creating the router instance to initialize it properly.
	Configure()
}
