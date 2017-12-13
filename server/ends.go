package server

import (
	"net/http"
	. "github.com/xuybin/go-api-gateway/types"
	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"strings"
)

func (s *GatewayServer) userAuth(c echo.Context) (err error) {
	sess := session.Default(c)
	user := User{}
	if err = c.Bind(&user); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if s.authUserService.AuthUser(user.Username, user.Password) {
		sess.Set(KEY_Username, user.Username)
		sess.Save()
		return c.String(http.StatusOK,"")
	} else {
		err=&ErrorMessage{ERR_PARAMETER, "login failed."}
		return
	}
}

func (s *GatewayServer) userRegister(c echo.Context) (err error) {
	user := &User{}
	if err = c.Bind(&user); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if err = s.authUserService.SaveUser(user.Username, user.Password); err == nil {
		s.AddRoleForUser(user.Username, s.DefaultRegisterRole)
		return c.String(http.StatusOK,"")
	} else {
		err=&ErrorMessage{ERR_PARAMETER, err.Error()}
		return
	}
}

func (s *GatewayServer) userUpdate(c echo.Context) (err error) {
	user := &User{}
	if err = c.Bind(&user); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if s.authUserService.UpdatePassword(user.Username, user.Password, user.NewPassword) {
		return c.String(http.StatusOK,"")
	} else {
		err=&ErrorMessage{ERR_PARAMETER, "update failed."}
		return
	}
}

func (s *GatewayServer) enforceAuth(c echo.Context) (err error) {
	p := new(Policy)
	if err = c.Bind(p); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if p.User == "" {
		p.User = KEY_CasbinAnonymous
	}
	if s.Enforce(p.User, p.Path, strings.ToUpper(p.Method)){
		return c.String(http.StatusOK,"")
	}else{
		err=&ErrorMessage{ERR_PARAMETER, "enforce auth failed."}
		return
	}
}

func (s *GatewayServer) addPolicy(c echo.Context) (err error) {
	p := new(Policy)
	if err = c.Bind(p); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if s.AddPolicy(p.User, p.Path, strings.ToUpper(p.Method)){
		return c.String(http.StatusOK,"")
	}else{
		err=&ErrorMessage{ERR_PARAMETER, "add policy failed."}
		return
	}
}

func (s *GatewayServer) removePolicy(c echo.Context) (err error) {
	p := new(Policy)
	if err = c.Bind(p); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if s.RemovePolicy(p.User, p.Path,  strings.ToUpper(p.Method)){
		return c.String(http.StatusOK,"")
	}else{
		err=&ErrorMessage{ERR_PARAMETER, "remove policy failed."}
		return
	}
}

func (s *GatewayServer) getPolicies(c echo.Context) (err error) {
	data := s.GetPolicy()
	policys:=[]Policy{}
	userParam:= c.QueryParam("user")
	pathParam:= c.QueryParam("path")
	for _,d :=range data{
		if len(d)>=3  && (userParam==""|| userParam==d[0]) && (pathParam==""|| pathParam==d[1]){
			policys=append(policys,Policy{User:d[0],Path:d[1],Method:d[2]})
		}
	}
	return c.JSON(http.StatusOK, policys)
}




func (s *GatewayServer) getGroupPolicies(c echo.Context) (err error) {
	data := s.GetGroupingPolicy()
	policys:=[]PolicyGroup{}
	userParam:= c.QueryParam("user")
	groupParam:= c.QueryParam("group")
	for _,d :=range data{
		if len(d)>=2 && (userParam==""|| userParam==d[0]) && (groupParam==""|| groupParam==d[1]){
			policys=append(policys, PolicyGroup{User:d[0], Group:d[1]})
		}
	}
	return c.JSON(http.StatusOK, policys)
}

func (s *GatewayServer) addGroupPolicy(c echo.Context) (err error) {
	pg := new(PolicyGroup)
	if err = c.Bind(pg); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	if s.AddRoleForUser(pg.User, pg.Group){
		return c.String(http.StatusOK,"")
	}else{
		err=&ErrorMessage{ERR_PARAMETER, "add groupPolicy failed."}
		return
	}
}

func (s *GatewayServer) removeRoleFromUser(c echo.Context) (err error) {
	pg := new(PolicyGroup)
	if err = c.Bind(pg); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	s.DeleteRoleForUser(pg.User, pg.Group)
	return c.String(http.StatusOK,"")
}


func (s *GatewayServer) upMetadata(c echo.Context) (err error) {
	if err = s.Enforcer.LoadPolicy(); err != nil {
		err=&ErrorMessage{ERR_PARAMETER, "parameter bind failed."}
		return
	}
	return c.String(http.StatusOK,"")
}

