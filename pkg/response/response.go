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
	Warnings []*any        `json:"warnings"`
	Errors   []*ErrorInner `json:"errors"`
}

// Success constructs a successful response with the provided data and warnings.
// It sends a JSON response with the specified status code and message.
func Success[T any](c *gin.Context, statusCode int, msg string, data *T, warnings []*any) {
	c.JSON(statusCode, Response[T]{
		Message:  msg,
		Data:     *data,
		Warnings: warnings,
		Errors:   nil,
	})
}

// Failure constructs an error response with the provided status code, error code, errors, and warnings.
// It sends a JSON response with the specified status code and message.
func Failure(c *gin.Context, statusCode int, errorCode ErrorCode, errors []*ErrorInner, warnings []*any) {
	c.AbortWithStatusJSON(statusCode, Response[any]{
		Message:  messages[errorCode],
		Data:     nil,
		Warnings: warnings,
		Errors:   errors,
	})
}

// NotImplemented constructs a response indicating that the requested functionality is not implemented.
// It sends a JSON response with a 501 Not Implemented status code and an appropriate error message.
func NotImplemented(c *gin.Context) {
	Failure(
		c,
		http.StatusNotImplemented,
		FATA000002,
		[]*ErrorInner{
			{
				Code:   string(FATA000002),
				Source: map[string]string{},
			},
		},
		nil,
	)
}

// DefaultValidatorError returns a default error response for validation errors.
// It contains a generic error message indicating an invalid request.
func DefaultValidatorError() []*ErrorInner {
	return []*ErrorInner{
		{
			Code: "",
			Source: map[string]string{
				"field":    "",
				"messages": "Invalid request",
			},
		},
	}
}

// TooManyRequest constructs a response indicating that the rate limit has been exceeded.
// It sends a JSON response with a 429 Too Many Requests status code and an appropriate error message.
func TooManyRequest(c *gin.Context) {
	Failure(
		c,
		http.StatusTooManyRequests,
		ESYS000010,
		[]*ErrorInner{
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

// AuthorizationHeaderError constructs a response indicating that the authorization header is missing.
// It sends a JSON response with a 401 Unauthorized status code and an appropriate error message.
func AuthorizationHeaderError(c *gin.Context) {
	Failure(
		c,
		http.StatusUnauthorized,
		ESYS000011,
		[]*ErrorInner{
			{Code: string(ESYS000011)},
		},
		nil,
	)
}

// Forbidden constructs a response indicating that access to the requested resource is forbidden.
// It sends a JSON response with a 403 Forbidden status code and an appropriate error message.
func Forbidden(c *gin.Context, defaultErrorCode ...ErrorCode) {
	errorCode := EBIZ000003

	if len(defaultErrorCode) > 0 {
		errorCode = defaultErrorCode[0]
	}

	Failure(
		c,
		http.StatusForbidden,
		errorCode,
		[]*ErrorInner{
			{
				Code: string(errorCode),
			},
		},
		nil,
	)
}

// HandleBizFailure constructs a business logic failure response.
// It sends a JSON response with the specified error code and an optional default status code.
// If the error code starts with "FATA", it sends a 500 Internal Server Error response.
func HandleBizFailure(c *gin.Context, code ErrorCode, defaultStatusCode ...int) {
	errorCode := string(code)
	errorMsg := messages[code]
	errorResponse := Response[any]{
		Message:  errorMsg,
		Data:     nil,
		Warnings: nil,
		Errors: []*ErrorInner{
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
