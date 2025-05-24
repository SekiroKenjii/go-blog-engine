package response

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrorInner struct {
	Code   string `json:"code"`
	Source any    `json:"source"`
}

type Response[T any] struct {
	Message  string        `json:"message"`
	Data     T             `json:"data"`
	Warnings []any         `json:"warnings"`
	Errors   *[]ErrorInner `json:"errors"`
}

func Success[T any](c *gin.Context, statusCode int, msg string, data T, warnings []any) {
	c.JSON(statusCode, Response[T]{
		Message:  msg,
		Data:     data,
		Warnings: warnings,
		Errors:   nil,
	})
}

func Failure(c *gin.Context, statusCode int, errorCode ErrorCode, errors *[]ErrorInner, warnings []any) {
	c.AbortWithStatusJSON(statusCode, Response[any]{
		Message:  messages[errorCode],
		Data:     nil,
		Warnings: warnings,
		Errors:   errors,
	})
}

func NotImplemented(c *gin.Context) {
	Failure(
		c,
		http.StatusNotImplemented,
		FATA000002,
		&[]ErrorInner{
			{
				Code:   string(FATA000002),
				Source: map[string]string{},
			},
		},
		nil,
	)
}

func DefaultValidatorError() *[]ErrorInner {
	return &[]ErrorInner{
		{
			Code: "",
			Source: map[string]string{
				"field":    "",
				"messages": "Invalid request",
			},
		},
	}
}

func TooManyRequest(c *gin.Context) {
	Failure(
		c,
		http.StatusTooManyRequests,
		ESYS000010,
		&[]ErrorInner{
			{
				Code: string(ESYS000010),
				Source: map[string]string{
					"ip":   c.ClientIP(),
					"path": c.Request.URL.Path,
				},
			},
		},
		nil,
	)
}

func AuthenticationHeaderError(c *gin.Context) {
	Failure(
		c,
		http.StatusUnauthorized,
		ESYS000011,
		&[]ErrorInner{
			{
				Code: string(ESYS000011),
			},
		},
		nil,
	)
}

func Forbidden(c *gin.Context) {
	Failure(
		c,
		http.StatusForbidden,
		EBIZ000003,
		&[]ErrorInner{
			{
				Code: string(EBIZ000003),
			},
		},
		nil,
	)
}

func HandleBizFailure(c *gin.Context, code ErrorCode, defaultStatusCode ...int) {
	errorCode := string(code)
	errorMsg := messages[code]
	errorResponse := Response[any]{
		Message:  errorMsg,
		Data:     nil,
		Warnings: nil,
		Errors: &[]ErrorInner{
			{
				Code:   errorCode,
				Source: map[string]string{"message": errorMsg},
			},
		},
	}

	if strings.HasPrefix(errorCode, "FATA") {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)

		return
	}

	if len(defaultStatusCode) > 0 {
		c.AbortWithStatusJSON(defaultStatusCode[0], errorResponse)

		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse)
}
