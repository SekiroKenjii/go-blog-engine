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

// RegisterPublicRoutes registers authentication routes that don't require authentication.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// Public routes - no authentication required
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.GET("/verify-email", h.VerifyEmail)
	rg.POST("/send-password-reset", h.SendPasswordReset)
	rg.POST("/verify-password-reset-token", h.VerifyPasswordResetToken)
	rg.POST("/reset-password", h.ResetPassword)
}

// RegisterProtectedRoutes registers authentication routes that require authentication.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	// Protected routes - authentication required
	rg.POST("/refresh-token", h.RefreshToken)
	rg.POST("/logout", h.Logout)
	rg.POST("/send-verification-email", h.SendVerificationEmail)
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

	userID, bizErrCode := h.service.Register(c.Request.Context(), req.Email, req.FirstName, req.LastName, req.Password)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)

		return
	}

	// don't response error if email sending fails
	// this is to ensure user registration is successful even if email verification fails
	// if email sending fails, user can request a new verification email later
	_ = h.service.SendVerificationEmail(c.Request.Context(), req.Email, userID)

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

	tokenPair, bizErrCode := h.service.Login(c.Request.Context(), req.Email, req.Password, deviceID, ip, ua)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)

		return
	}

	response.Success(
		c,
		http.StatusOK,
		"User logged in successfully",
		&AuthResponse{
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
	userID := accessor.GetUserID(c)
	if userID == "" {
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
		&AuthResponse{
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
	userID := accessor.GetUserID(c)
	if userID == "" {
		response.Forbidden(c)

		return
	}

	var req LogoutRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)

		return
	}

	deviceID, _, _ := accessor.GetDeviceInfo(c)

	bizErrCode := h.service.Logout(c.Request.Context(), userID, deviceID, req.RefreshToken)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode, http.StatusUnauthorized)

		return
	}

	response.Success[any](c, http.StatusOK, "User logged out successfully", nil, nil)
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
// @Router      /api/v1/auth/verify-email [get]
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := accessor.GetQueryParam(c, "token")
	if token == "" {
		// Try to get from request body if not in query params
		var req VerifyEmailRequest
		if err := validator.ValidateRequest(c, &req); err == nil {
			token = req.Token
		}
	}

	if token == "" {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000006, []*response.ErrorInner{
			{Code: string(response.EBIZ000006), Source: "Token is required"},
		}, nil)
		return
	}

	bizErrCode := h.service.VerifyEmail(c.Request.Context(), token)
	if bizErrCode == response.FATA001001 || bizErrCode == response.FATA002001 {
		response.HandleBizFailure(c, bizErrCode, http.StatusInternalServerError)
		return
	}

	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)
		return
	}

	response.Success[any](c, http.StatusOK, "Email verified successfully", nil, nil)
}

// SendVerificationEmail godoc
// @Summary     Send verification email
// @Description Send verification email to user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Verification email sent successfully"
// @Failure     400 {object} response.Response[any] "Validation error"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/send-verification-email [post]
func (h *Handler) SendVerificationEmail(c *gin.Context) {
	var req SendVerificationEmailRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)
		return
	}

	// Get user by email to get user ID
	user, err := h.service.(*AuthService).repo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ001000, []*response.ErrorInner{
			{Code: string(response.EBIZ001000), Source: "User not found"},
		}, nil)
		return
	}

	bizErrCode := h.service.SendVerificationEmail(c.Request.Context(), req.Email, user.ID)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)
		return
	}

	response.Success[any](c, http.StatusOK, "Verification email sent successfully", nil, nil)
}

// SendPasswordReset godoc
// @Summary     Send password reset email
// @Description Send password reset email to user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Password reset email sent successfully"
// @Failure     400 {object} response.Response[any] "Validation error"
// @Failure     500 {object} response.Response[any] "Internal server error"
// @Router      /api/v1/auth/send-password-reset [post]
func (h *Handler) SendPasswordReset(c *gin.Context) {
	var req SendPasswordResetRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)
		return
	}

	bizErrCode := h.service.SendPasswordResetEmail(c.Request.Context(), req.Email)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)
		return
	}

	response.Success[any](c, http.StatusOK, "Password reset email sent successfully", nil, nil)
}

// VerifyPasswordResetToken godoc
// @Summary     Verify password reset token
// @Description Verify password reset token validity
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[any] "Token is valid"
// @Failure     400 {object} response.Response[any] "Validation error or invalid token"
// @Router      /api/v1/auth/verify-password-reset-token [post]
func (h *Handler) VerifyPasswordResetToken(c *gin.Context) {
	var req VerifyPasswordResetTokenRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)
		return
	}

	bizErrCode := h.service.VerifyPasswordResetToken(c.Request.Context(), req.Email, req.Token)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)
		return
	}

	response.Success[any](c, http.StatusOK, "Token is valid", nil, nil)
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
	var req ResetPasswordRequest

	if err := validator.ValidateRequest(c, &req); err != nil {
		response.Failure(c, http.StatusBadRequest, response.EBIZ000002, err, nil)
		return
	}

	bizErrCode := h.service.ResetPassword(c.Request.Context(), req.Email, req.NewPassword, req.Token)
	if bizErrCode != response.SBIZ000001 {
		response.HandleBizFailure(c, bizErrCode)
		return
	}

	response.Success[any](c, http.StatusOK, "Password reset successfully", nil, nil)
}
