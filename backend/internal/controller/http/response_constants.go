package httpapi

const (
	problemContentType              = "application/problem+json"
	problemTitleUnauthorized        = "Unauthorized"
	problemTitleInternalServer      = "Internal Server Error"
	problemCodeAuthRequired         = "AUTH_REQUIRED"
	problemCodeCategoryLoadFailed   = "CATEGORY_LOAD_FAILED"
	problemCodeAuthInvalidFormat    = "AUTH_INVALID_FORMAT"
	problemCodeTokenInvalid         = "TOKEN_INVALID"
	problemCodeSessionExpired       = "SESSION_EXPIRED"
	problemCodeSessionVerifyFailed  = "SESSION_VERIFY_FAILED"
	problemCodeUserNotFound         = "USER_NOT_FOUND"
	problemCodeProfileLoadFailed    = "PROFILE_LOAD_FAILED"
	problemCodeHomeLoadFailed       = "HOME_LOAD_FAILED"
	problemCodeBadRequest           = "BAD_REQUEST"
	problemCodeLoginFailed          = "LOGIN_FAILED"
	problemCodeGoogleUnavailable    = "GOOGLE_LOGIN_UNAVAILABLE"
	problemCodeGoogleStateInvalid   = "GOOGLE_STATE_INVALID"
	problemCodeGoogleProfileInvalid = "GOOGLE_PROFILE_INVALID"
	problemCodeEmailNotAllowed      = "EMAIL_NOT_ALLOWED"
)
