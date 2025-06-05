package validator

import (
	"errors"

	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/gin-gonic/gin"

	goValidator "github.com/go-playground/validator/v10"
)

type FieldErrorCallback func(f goValidator.FieldError) (string, string)

// ValidateRequest validates the incoming JSON request body against the provided struct type T.
// It returns a slice of validation errors if any validation fails, or nil if the request is valid.
func ValidateRequest[T any](c *gin.Context, req *T) []*response.ErrorInner {
	if err := c.ShouldBindJSON(&req); err != nil {
		var result []*response.ErrorInner
		var val goValidator.ValidationErrors

		if errors.As(err, &val) {
			for _, f := range val {
				msg, code := getValidateMsgCode(f)

				result = append(result, &response.ErrorInner{
					Code: string(code),
					Source: map[string]string{
						"field":    f.Field(),
						"messages": utils.FormatFieldError(msg, f),
					},
				})
			}

			return result
		}

		return response.DefaultValidatorError()
	}

	return nil
}

// getValidateMsgCode retrieves the validation message and error code for a given field error.
func getValidateMsgCode(f goValidator.FieldError) (string, response.ErrorCode) {
	tag := f.Tag()

	message := MessageCodes[ValidateErrorCode(tag)]

	if message == nil {
		message = MessageCodes[ErrDefault]
	}

	return message.Message, message.Code
}
