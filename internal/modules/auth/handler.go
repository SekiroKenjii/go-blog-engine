package auth

import (
	"net/http"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/pkg/accessor"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/validator"
	"github.com/gin-gonic/gin"

	_ "github.com/SekiroKenjii/go-blog-engine/docs"
)

var (
	handler     IAuthHandler
	handlerOnce sync.Once
)

type Handler struct {
	Service IAuthService
}

func Instance() IAuthHandler {
	handlerOnce.Do(func() {
		handler = &Handler{Service: NewService()}
	})

	return handler
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh-token", h.Refresh)
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

	if bizErrCode := h.Service.Register(c.Request.Context(), req); bizErrCode != response.SBIZ000001 {
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

	tokenPair, bizErrCode := h.Service.Login(c.Request.Context(), req, deviceID, ip, ua)
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

// Refresh godoc
// @Summary     Refresh user token
// @Description Refresh user access token using refresh token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[TokenPair] "Token refreshed successfully"
// @Failure     401 {object} response.Response[any] "Invalid access token or refresh token or token expired"
// @Failure     500 {object} response.Response[any] "Internal server error: cryptographic operation failed"
// @Router      /api/v1/auth/refresh-token [post]
func (h *Handler) Refresh(c *gin.Context) {
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

	tokenPair, bizErrCode := h.Service.RefreshToken(c.Request.Context(), userID, req.RefreshToken)
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
