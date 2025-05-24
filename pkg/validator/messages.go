package validator

import "github.com/SekiroKenjii/go-blog-engine/pkg/response"

type ValidateMessageCode struct {
	Message string
	Code    response.ErrorCode
}

type ValidateErrorCode string

const (
	ErrRequiredField ValidateErrorCode = "required"
	ErrInvalidEmail  ValidateErrorCode = "email"
	ErrMinLength     ValidateErrorCode = "min"
	ErrDefault       ValidateErrorCode = ""
)

var MessageCodes = map[ValidateErrorCode]*ValidateMessageCode{
	ErrRequiredField: {Message: "Field \"__FIELD__\" is required", Code: response.EBIZ000004},
	ErrInvalidEmail:  {Message: "Field \"__FIELD__\" must be a valid email address", Code: response.EBIZ000005},
	ErrMinLength:     {Message: "Field \"__FIELD__\" must be at least __PARAM__ characters long", Code: response.EBIZ000005},
	ErrDefault:       {Message: "Field \"__FIELD__\": __PARAM__", Code: response.EBIZ000005},
}
