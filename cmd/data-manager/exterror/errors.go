package exterror

//common error
var (
	ErrInterval         = NewError(ErrCodeInterval, "interval error")
	ErrParameterInvalid = NewError(ErrCodeParameterInvalid, "parameter invalid error")
	ErrUnauthorized     = NewError(ErrCodeUnauthorized, "unauthorized error")
)
