package types

const  KEY_Username = "username"
const  KEY_CasbinAnonymous = "casbin_anonymous"
const  KEY_BasicRole = "basic_role"
const  ERR_UNAUTHORIZED = "err_unauthorized"
const  ERR_FORBIDDEN = "err_forbidden"
const ERR_PARAMETER = "err_parameter"

func ErrUnauthorized() *ErrorMessage {
	return &ErrorMessage{ ERR_UNAUTHORIZED, "unauthenticated, please login."}
}
func  ErrForbidden() *ErrorMessage {
	return &ErrorMessage{ ERR_FORBIDDEN, "not authorized, please authorization."}
}
func  ErrParameter() *ErrorMessage {
	return &ErrorMessage{ ERR_PARAMETER, "parameter error, please check and try again."}
}