package server

import (
	. "github.com/xuybin/go-api-gateway/types"
	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
)

// BasicAuthSessionMw is used for reading basic auth header and save username if it passed
func (s *GatewayServer) BasicAuthSessionMw(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username, password, exist := c.Request().BasicAuth()
		sess := session.Default(c)
		if exist {
			if(username != "" && s.authUserService.AuthUser(username, password)){
				sess.Set(KEY_Username, username)
			}else {
				sess.Delete(KEY_Username)
			}
			sess.Save()
		}
		return next(c)
	}
}
