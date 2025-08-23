package response

type ErrorCode string

const (
	// #region Internal Error Codes

	// Error: Success Code
	SBIZ000001 ErrorCode = "SBIZ000001"

	// Error: Invalid Request
	EBIZ000001 ErrorCode = "EBIZ000001"
	// Error: Validation Failed
	EBIZ000002 ErrorCode = "EBIZ000002"
	// Error: Authentication Required
	EBIZ000003 ErrorCode = "EBIZ000003"
	// Error: Request Body Field Required
	EBIZ000004 ErrorCode = "EBIZ000004"
	// Error: Request Body Field Invalid
	EBIZ000005 ErrorCode = "EBIZ000005"
	// Error: Request Parameters Required
	EBIZ000006 ErrorCode = "EBIZ000006"
	// Error: Request Parameters Invalid
	EBIZ000007 ErrorCode = "EBIZ000007"
	// Error: Wrong User Credentials
	EBIZ001000 ErrorCode = "EBIZ001000"
	// Error: Wrong User Password
	EBIZ001001 ErrorCode = "EBIZ001001"
	// Error: Invalid User Access Token
	EBIZ001002 ErrorCode = "EBIZ001002"
	// Error: Invalid User Refresh Token
	EBIZ001003 ErrorCode = "EBIZ001003"
	// Error: User Refresh Token Expired or unauthenticated user
	EBIZ001004 ErrorCode = "EBIZ001004"
	// Error: Email verification token invalid or expired
	EBIZ001005 ErrorCode = "EBIZ001005"
	// Error: Password reset token invalid or expired
	EBIZ001006 ErrorCode = "EBIZ001006"
	// Error: User account not verified
	EBIZ001007 ErrorCode = "EBIZ001007"
	// Error: User account is locked
	EBIZ001008 ErrorCode = "EBIZ001008"

	// Fatal: Rate Limit Exceeded
	ESYS000010 ErrorCode = "ESYS000010"
	// Fatal: Authentication Header Not Found
	ESYS000011 ErrorCode = "ESYS000011"

	// Warning: Server Warning
	WBIZ000001 ErrorCode = "WBIZ000001"

	// Fatal: Internal Server Error
	FATA000001 ErrorCode = "FATA000001"
	// Fatal: API Not Implemented
	FATA000002 ErrorCode = "FATA000002"
	// Fatal: Cryptographic Operation Error
	FATA000101 ErrorCode = "FATA000101"
	// Fatal: Write Database Error
	FATA001001 ErrorCode = "FATA001001"
	// Fatal: Write Cache Error
	FATA002001 ErrorCode = "FATA002001"

	// #endregion

	// #region External Error Codes
	/** Define error codes for external services here */
	// #endregion External Error Code
)

const (
	invalidRequest          = "Invalid Request"
	validationErrorOccurred = "One or more errors occurred while validating the request"
	authenticationRequired  = "Authentication Required"
	invalidUserCredentials  = "Invalid User Credentials"
	rateLimitExceeded       = "Rate Limit Exceeded"
	internalServerError     = "Internal Server Error"
	apiNotImplemented       = "API Not Implemented"
	internalServerWarning   = "Internal Server Warning"
)

var messages = map[ErrorCode]string{
	// #region: Internal Error Messages

	EBIZ000001: invalidRequest,
	EBIZ000002: validationErrorOccurred,
	EBIZ000003: authenticationRequired,
	EBIZ000004: invalidRequest,
	EBIZ000005: invalidRequest,
	EBIZ000006: invalidRequest,
	EBIZ000007: invalidRequest,
	EBIZ001000: invalidUserCredentials,
	EBIZ001001: invalidUserCredentials,
	EBIZ001002: invalidUserCredentials,
	EBIZ001003: invalidUserCredentials,
	EBIZ001004: invalidUserCredentials,
	EBIZ001005: invalidUserCredentials,
	EBIZ001006: invalidUserCredentials,
	EBIZ001007: invalidUserCredentials,
	EBIZ001008: invalidUserCredentials,

	ESYS000010: rateLimitExceeded,
	ESYS000011: authenticationRequired,

	WBIZ000001: internalServerWarning,

	FATA000001: internalServerError,
	FATA000002: apiNotImplemented,
	FATA000101: internalServerError,
	FATA001001: internalServerError,
	FATA002001: internalServerError,

	// #endregion

	// #region: External Error Messages
	/** Define error messages for external services here */
	// #endregion External Error Messages
}
