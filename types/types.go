package types

import "fmt"

type Policy struct {
	User   string `json:"user" form:"user" query:"user"`
	Path   string `json:"path" form:"path" query:"path"`
	Method string `json:"method" form:"method" query:"method"`
}

type PolicyGroup struct {
	User  string `json:"user" form:"user" query:"user"`
	Group string `json:"group" form:"group" query:"group"`
}

type User struct {
	Username    string `json:"username" form:"username" query:"username"`
	Password    string `json:"password" form:"password" query:"password"`
	NewPassword string `json:"new_password" form:"new_password" query:"new_password"`
}

// ErrorMessage
type ErrorMessage struct {
	ErrorTitle string      `json:"error" form:"error" query:"error"`
	ErrorDescription    string `json:"error_description"  form:"error_description" query:"error_description"`
	//PrettyDescription   string `json:"pretty_description"  form:"pretty_description" query:"pretty_description"`
}
// Error makes it compatible with `error` interface.
func (em *ErrorMessage) Error() string {
	return fmt.Sprintf("error=%s, error_description=%s", em.ErrorTitle, em.ErrorDescription)
}
