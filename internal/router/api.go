package router

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/auth"
	"github.com/gin-gonic/gin"
)

// addAPIRoutes adds the API routes to the gin engine.
// It creates a new group for API routes and registers the versioned routes under it.
// The API routes are grouped under "/api" and further divided into versioned groups like "/v1".
// This structure allows for easy expansion of the API in the future, supporting multiple versions.
func (r *Router) addAPIRoutes() {
	apiRoutes := r.engine.Group("/api")

	r.addAPIRoutesV1(apiRoutes)
}

// addAPIRoutesV1 adds the version 1 API routes to the gin engine.
func (r *Router) addAPIRoutesV1(apiRoutes *gin.RouterGroup) {
	v1 := apiRoutes.Group("/v1")
	{
		register(auth.NewAuthHandler(), v1)
	}
}

// register registers the given handler's routes under the specified router group.
// It calls the RegisterRoutes method of the handler, passing the router group to it.
func register(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterRoutes(rg)
}
