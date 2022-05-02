package error_constant

const (
	LoginInvalid       = "your username or password is not valid"
	LoginFailed        = "something went wrong, please try to logging in again later"
	RegisterInvalid    = "email already exists"
	RegisterFailed     = "something went wrong, please try to register again later"
	VerifyTokenInvalid = "session is empty or invalid"
	VerifyTokenFailed  = "something went wrong, please try to re-login"
)
