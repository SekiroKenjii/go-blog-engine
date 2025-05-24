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
	// Error: Wrong User Credentials
	EBIZ001000 ErrorCode = "EBIZ001000"
	// Error: Wrong User Password
	EBIZ001001 ErrorCode = "EBIZ001001"
	// Error: Invalid User Access Token
	EBIZ001002 ErrorCode = "EBIZ001002"
	// Error: Invalid User Refresh Token
	EBIZ001003 ErrorCode = "EBIZ001003"
	// Error: User Refresh Token Expired
	EBIZ001004 ErrorCode = "EBIZ001004"
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

	// #endregion

	// #region External Error Codes
	/** Define error codes for external services here */
	// #endregion External Error Code
)

var messages = map[ErrorCode]string{
	// #region: Internal Error Messages

	EBIZ000001: "Invalid Request",
	EBIZ000002: "One or more errors occurred while validating the request",
	EBIZ000003: "Authentication Required",
	EBIZ000004: "",
	EBIZ000005: "",
	EBIZ001000: "Wrong User Credentials",
	EBIZ001001: "Wrong User Credentials",
	EBIZ001002: "Invalid User Credentials",
	EBIZ001003: "Invalid User Credentials",
	EBIZ001004: "Invalid User Credentials",
	ESYS000010: "Rate Limit Exceeded",
	ESYS000011: "Authentication Header Not Found",

	WBIZ000001: "Server Warning",

	FATA000001: "Internal Server Error",
	FATA000002: "API Not Implemented",
	FATA000101: "Internal Server Error",
	FATA001001: "Internal Server Error",

	// #endregion

	// #region: External Error Messages
	/** Define error messages for external services here */
	// #endregion External Error Messages
}
