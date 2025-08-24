package router

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/middlewares"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/auth"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/posts"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/users"
	"github.com/gin-gonic/gin"
)

// addAPIRoutes adds the API routes to the gin engine.
// It creates separate groups for public and protected API routes.
// Public routes don't require authentication, while protected routes do.
// This architecture allows each module to define which routes are public vs protected.
func (r *Router) addAPIRoutes() {
	apiRoutes := r.engine.Group("/api")

	r.addAPIRoutesV1(apiRoutes)
}

// addAPIRoutesV1 adds the version 1 API routes to the gin engine.
// Routes are organized into public and protected groups for better security control.
func (r *Router) addAPIRoutesV1(apiRoutes *gin.RouterGroup) {
	v1 := apiRoutes.Group("/v1")

	publicRoutes := v1.Group("")
	{
		authPublicRoutes := publicRoutes.Group("/auth")
		authPublicRoutes.Use(middlewares.AuthRateLimit())
		{
			registerPublicRoutes(auth.NewAuthHandler(), authPublicRoutes)
		}

		postsPublicRoutes := publicRoutes.Group("/posts")
		{
			registerPublicRoutes(posts.NewPostHandler(), postsPublicRoutes)
		}

		usersPublicRoutes := publicRoutes.Group("/users")
		{
			registerPublicRoutes(users.NewUserHandler(), usersPublicRoutes)
		}
	}

	protectedRoutes := v1.Group("")
	protectedRoutes.Use(middlewares.RequireAuth())
	{
		authProtectedRoutes := protectedRoutes.Group("/auth")
		{
			registerProtectedRoutes(auth.NewAuthHandler(), authProtectedRoutes)
		}

		postsProtectedRoutes := protectedRoutes.Group("/posts")
		{
			registerProtectedRoutes(posts.NewPostHandler(), postsProtectedRoutes)
		}

		usersProtectedRoutes := protectedRoutes.Group("/users")
		{
			registerProtectedRoutes(users.NewUserHandler(), usersProtectedRoutes)
		}
	}
}

// registerPublicRoutes registers public routes that don't require authentication
func registerPublicRoutes(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterPublicRoutes(rg)
}

// registerProtectedRoutes registers protected routes that require authentication
func registerProtectedRoutes(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterProtectedRoutes(rg)
}
