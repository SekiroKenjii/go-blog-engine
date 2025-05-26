package auth

import (
	"net/http"

	"github.com/SekiroKenjii/go-blog-engine/pkg/accessor"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/validator"
	"github.com/gin-gonic/gin"

	_ "github.com/SekiroKenjii/go-blog-engine/docs"
)

type Handler struct {
	service IAuthService
}

func NewAuthHandler() IAuthHandler {
	return &Handler{service: NewAuthService()}
}

// RegisterRoutes implements IAuthHandler.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh-token", h.RefreshToken)
	auth.POST("/logout", h.Logout)
	auth.POST("/verify-email", h.VerifyEmail)
	auth.POST("/verify-phone", h.VerifyPhone)
	auth.POST("/resend-verification-email", h.ResendVerificationEmail)
	auth.POST("/resend-verification-phone", h.ResendVerificationPhone)
	auth.POST("/send-password-reset-email", h.SendPasswordResetEmail)
	auth.POST("/verify-password-reset-token", h.VerifyPasswordResetToken)
	auth.POST("/reset-password", h.ResetPassword)
}

// Register godoc
// @Summary 	Register new user
// @Description Create a new user account
// @Tags 		auth
// @Success 	201 {object} response.Response[any] "User registered successfully"
// @Failure 	400 {object} response.Response[any] "Validation error or registration failed"
// @Failure 	500 {object} response.Response[any] "Internal server error"
// @Router 		/api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)

		return
	}

	if bizErrCode := h.service.Register(c.Request.Context(), req); bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)

		return
	}

	response.Success[any](c, http.StatusCreated, "User registered successfully", nil, nil)
}

// Login godoc
// @Summary     Login user
// @Description Authenticate user and return token pair
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[TokenPair] "User logged in successfully"
// @Failure     400 {object} response.Response[any] "Validation error or login failed"
// @Failure     500 {object} response.Response[any] "Internal server error: cryptographic operation failed"
// @Router      /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)

		return
	}

	deviceID, ip, ua := accessor.GetDeviceInfo(c)

	tokenPair, bizErrCode := h.service.Login(c.Request.Context(), req, deviceID, ip, ua)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)

		return
	}

	response.Success(
		c,
		http.StatusOK,
		"User logged in successfully",
		AuthResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
		nil,
	)
}

// RefreshToken godoc
// @Summary     Refresh user token
// @Description Refresh user access token using refresh token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[TokenPair] "Token refreshed successfully"
// @Failure     401 {object} response.Response[any] "Invalid access token or refresh token or token expired"
// @Failure     500 {object} response.Response[any] "Internal server error: cryptographic operation failed"
// @Router      /api/v1/auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	userID, exists := accessor.GetUserID(c)
	if !exists {
		response.Forbidden(c)

		return
	}

	var req RefreshRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)

		return
	}

	tokenPair, bizErrCode := h.service.RefreshToken(c.Request.Context(), userID, req.RefreshToken)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode, http.StatusUnauthorized)

		return
	}

	response.Success(
		c,
		http.StatusOK,
		"Token refreshed successfully",
		AuthResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
		nil,
	)
}

// Logout godoc
// @Summary     Logout user
// @Description Logout user and invalidate tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "User logged out successfully"
// @Failure     401 {object} response.Response[any] "Invalid access token or user not authenticated"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// TODO: Implement Logout logic
}

// VerifyEmail godoc
// @Summary     Verify user email
// @Description Verify user email address using verification code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Email verified successfully"
// @Failure     400 {object} response.Response[any] "Validation error or email verification failed"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	// TODO: Implement VerifyEmail logic
}

// VerifyPhone godoc
// @Summary     Verify user phone number
// @Description Verify user phone number using verification code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Phone number verified successfully"
// @Failure     400 {object} response.Response[any] "Validation error or phone verification failed"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/verify-phone [post]
func (h *Handler) VerifyPhone(c *gin.Context) {
	// TODO: Implement VerifyPhone logic
}

// ResendVerificationEmail godoc
// @Summary     Resend verification email
// @Description Resend verification email to user if it was not received or expired
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Verification email resent successfully"
// @Failure     400 {object} response.Response[any] "Validation error or email not found"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/resend-verification-email [post]
func (h *Handler) ResendVerificationEmail(c *gin.Context) {
	// TODO: Implement ResendVerificationEmail logic
}

// ResendVerificationPhone godoc
// @Summary     Resend verification phone
// @Description Resend verification phone number to user if it was not received or expired
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Verification phone resent successfully"
// @Failure     400 {object} response.Response[any] "Validation error or phone not found"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/resend-verification-phone [post]
func (h *Handler) ResendVerificationPhone(c *gin.Context) {
	// TODO: Implement ResendVerificationPhone logic
}

// SendPasswordResetEmail godoc
// @Summary     Send password reset email
// @Description Send password reset email to user with reset link
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Password reset email sent successfully"
// @Failure     400 {object} response.Response[any] "Validation error or email not found"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/send-password-reset-email [post]
func (h *Handler) SendPasswordResetEmail(c *gin.Context) {
	// TODO: Implement SendPasswordResetEmail logic
}

// VerifyPasswordResetToken godoc
// @Summary     Verify password reset token
// @Description Verify password reset token to ensure it is valid and not expired
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Token verified successfully"
// @Failure     400 {object} response.Response[any] "Validation error or token invalid"
// @Failure     401 {object} response.Response[any] "Invalid reset token or token expired"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/verify-password-reset-token [post]
func (h *Handler) VerifyPasswordResetToken(c *gin.Context) {
	// TODO: Implement VerifyPasswordResetToken logic
}

// ResetPassword godoc
// @Summary     Reset user password
// @Description Reset user password using valid reset token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Password reset successfully"
// @Failure     400 {object} response.Response[any] "Validation error or reset failed"
// @Failure     401 {object} response.Response[any] "Invalid reset token or token expired"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	// TODO: Implement ResetPassword logic
}
