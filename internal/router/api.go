package router

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/auth"
	"github.com/gin-gonic/gin"
)

func (r *Router) addAPIRoutes() {
	apiRoutes := r.engine.Group("/api")

	r.addAPIRoutesV1(apiRoutes)
}

func (r *Router) addAPIRoutesV1(apiRoutes *gin.RouterGroup) {
	v1 := apiRoutes.Group("/v1")
	{
		register(auth.NewAuthHandler(), v1)
	}
}

func register(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterRoutes(rg)
}
