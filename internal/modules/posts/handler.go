package posts

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	// service IPostService // TODO: Implement when posts service is ready
}

func NewPostHandler() abstract.IHandler {
	return &Handler{}
}

// RegisterPublicRoutes implements abstract.IHandler.
// Registers post routes that don't require authentication.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// Public post routes - no authentication required
	// TODO: Implement when posts service is ready
	// rg.GET("", h.GetPosts)      // Get all posts
	// rg.GET("/:id", h.GetPost)   // Get single post by ID
}

// RegisterProtectedRoutes implements abstract.IHandler.
// Registers post routes that require authentication.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	// Protected post routes - authentication required
	// TODO: Implement when posts service is ready
	// rg.POST("", h.CreatePost)        // Create new post
	// rg.PUT("/:id", h.UpdatePost)     // Update existing post
	// rg.DELETE("/:id", h.DeletePost)  // Delete post
}
