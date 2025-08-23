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

	// Public routes - no authentication required
	publicRoutes := v1.Group("")
	{
		// Auth public routes with specific rate limiting (overrides global rate limiting)
		authPublicRoutes := publicRoutes.Group("/auth")
		// Note: AuthRateLimit is more restrictive than global RateLimit, so we skip global for auth routes
		authPublicRoutes.Use(middlewares.AuthRateLimit())
		{
			registerPublicRoutes(auth.NewAuthHandler(), authPublicRoutes)
		}

		// Posts public routes (uses global rate limiting from router.go)
		postsPublicRoutes := publicRoutes.Group("/posts")
		{
			registerPublicRoutes(posts.NewPostHandler(), postsPublicRoutes)
		}

		// Users public routes (uses global rate limiting from router.go)
		usersPublicRoutes := publicRoutes.Group("/users")
		{
			registerPublicRoutes(users.NewUserHandler(), usersPublicRoutes)
		}
	}

	// Protected routes - authentication required
	protectedRoutes := v1.Group("")
	protectedRoutes.Use(middlewares.RequireAuth())
	{
		// Auth protected routes (uses global rate limiting, no additional auth rate limiting needed)
		authProtectedRoutes := protectedRoutes.Group("/auth")
		{
			registerProtectedRoutes(auth.NewAuthHandler(), authProtectedRoutes)
		}

		// Posts protected routes (uses global rate limiting from router.go)
		postsProtectedRoutes := protectedRoutes.Group("/posts")
		{
			registerProtectedRoutes(posts.NewPostHandler(), postsProtectedRoutes)
		}

		// Users protected routes (uses global rate limiting from router.go)
		usersProtectedRoutes := protectedRoutes.Group("/users")
		{
			registerProtectedRoutes(users.NewUserHandler(), usersProtectedRoutes)
		}
	}
} // registerPublicRoutes registers public routes that don't require authentication
func registerPublicRoutes(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterPublicRoutes(rg)
}

// registerProtectedRoutes registers protected routes that require authentication
func registerProtectedRoutes(handler abstract.IHandler, rg *gin.RouterGroup) {
	handler.RegisterProtectedRoutes(rg)
}
