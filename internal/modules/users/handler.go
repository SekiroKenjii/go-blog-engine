package users

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	// service IUserService // TODO: Implement when users service is ready
}

func NewUserHandler() abstract.IHandler {
	return &Handler{}
}

// RegisterPublicRoutes implements abstract.IHandler.
// Registers user routes that don't require authentication.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// Public user routes - no authentication required
	// TODO: Implement when users service is ready
	// Most user routes will be protected, but some might be public like:
	// rg.GET("/:id/public-profile", h.GetPublicProfile) // Get public user info
}

// RegisterProtectedRoutes implements abstract.IHandler.
// Registers user routes that require authentication.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	// Protected user routes - authentication required
	// TODO: Implement when users service is ready
	// rg.GET("/profile", h.GetProfile)           // Get current user profile
	// rg.PUT("/profile", h.UpdateProfile)        // Update user profile
	// rg.DELETE("/profile", h.DeleteAccount)     // Delete user account
	// rg.POST("/change-password", h.ChangePassword) // Change password
}
