package server

import (
	"net/url"
	"github.com/go-openapi/spec"
	"github.com/xuybin/go-api-gateway/enforcer"
	. "github.com/xuybin/go-api-gateway/types"
	"github.com/xuybin/go-api-gateway/user"
	"github.com/casbin/casbin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"fmt"
	"github.com/go-openapi/jsonreference"
	"reflect"
)

type GatewayServer struct {
	*echo.Echo                            // web service
	*casbin.Enforcer                      // authorization service
	resourceHost        *url.URL          // be protected http resource
	authUserService     *user.UserService // user authenticate service
	DefaultRegisterRole string            // Default New User Group
	swaggerJSON *spec.Swagger
}

// NewGatewayServer instance
func NewGatewayServer(connStr string, resourceHostStr string, defaultRole ...string) (s *GatewayServer) {
	resourceHost, err := url.Parse(resourceHostStr)
	if err != nil {
		panic(err)
	}
	// construct gateway
	s = &GatewayServer{
		Echo:            echo.New(),
		Enforcer:        enforcer.NewCasbinEnforcer(connStr),
		resourceHost:    resourceHost,
		authUserService: user.NewUserService(connStr),
	}

	if len(defaultRole) == 1 {
		s.DefaultRegisterRole = defaultRole[0]
	} else {
		s.DefaultRegisterRole = KEY_BasicRole
	}
	s.HTTPErrorHandler=func (err error, c echo.Context) {
		if reflect.TypeOf(err) == reflect.TypeOf(&echo.HTTPError{}) {
			httpError := err.(*echo.HTTPError)
			c.JSON(httpError.Code, httpError.Message)
		}else if reflect.TypeOf(err) == reflect.TypeOf(&ErrorMessage{}) {
			errorMessage := err.(*ErrorMessage)
			if errorMessage.ErrorTitle==ERR_UNAUTHORIZED{
				c.JSON(http.StatusUnauthorized, errorMessage)
			}else if errorMessage.ErrorTitle==ERR_FORBIDDEN{
				c.JSON(http.StatusForbidden, errorMessage)
			}else if errorMessage.ErrorTitle==ERR_PARAMETER{
				c.JSON(http.StatusBadRequest, errorMessage)
			}
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"unknown_error", err.Error()})
		}
	}
	s.Use(NewCoockieSession())
	s.Static("/gateway/docs", "docs")
	s.swaggerJSON=initSwaggerJSON()
	s.mountAuthorizationEndPoints()
	s.mountReverseProxy()
	// hide echo banner
	s.Echo.HideBanner = true
	// load casbin policy from db
	s.Enforcer.LoadPolicy()
	return
}

func (s *GatewayServer) mountReverseProxy() {
	s.Group("/").Use(s.BasicAuthSessionMw, enforcer.Middleware(s.Enforcer), middleware.Proxy(&middleware.RoundRobinBalancer{
		Targets: []*middleware.ProxyTarget{
			&middleware.ProxyTarget{
				URL: s.resourceHost,
			},
		},
	}))
}

func (s *GatewayServer) mountAuthorizationEndPoints() {
	//s.File("/favicon.ico","docs/favicon-16x16.png")
	s.GET("/gateway/swagger/", s.getSwaggerJSON()).Name = "Swagger Infomation"

	s.POST("/auth/register/", s.userRegister).Name = "Register New User"
	s.POST("/auth/login/",s.userAuth).Name = "User Auth"
	s.PUT("/auth/password/", s.userUpdate).Name = "Passwrod Update"
	//验证策略(是否能验证组策略)
	s.POST("/auth/policy/", s.enforceAuth).Name = "Find Some Authority"

	//受控访问
	policy:=s.Group("/policy")
	policy.Use(enforcer.Middleware(s.Enforcer))
	//权限策略
	policy.GET("/", s.getPolicies).Name = "Get All Policies"
	policy.PUT("/", s.addPolicy).Name = "Add Policy"
	policy.DELETE("/", s.removePolicy).Name = "Remove Authority"

	//组策略(以用户,或组 未读查找)
	policy.GET("/group/", s.getGroupPolicies).Name = "Get Group Policies"
	policy.PUT("/group/", s.addGroupPolicy).Name = "Add Group To User"
	policy.DELETE("/group/", s.removeRoleFromUser).Name = "Remove Group From User"

	policy.HEAD("/metadata/", s.upMetadata).Name = "Remove Group From User"
}

var userDefinitionModel="auth"
var userGroupDefinitionModel ="policy_group"
var policyDefinitionModel="policy"
var metadataTag="metadata"
var authTag="auth"
var policyTag="policy"
var policyGroupTag="policy_group"
func initSwaggerJSON() (s *spec.Swagger){
	s = &spec.Swagger{}
	s.SwaggerProps = spec.SwaggerProps{}
	s.Swagger = "2.0"
	s.Schemes = []string{"http"}
	s.Tags =[]spec.Tag{
		{TagProps: spec.TagProps{Name: authTag, Description: "认证"}},
		{TagProps: spec.TagProps{Name: policyTag, Description: "权限策略"}},
		{TagProps: spec.TagProps{Name: policyGroupTag, Description: "组策略"}},
		}
	s.Info = &spec.Info{spec.VendorExtensible{}, spec.InfoProps{
		Title:       fmt.Sprintf("API网关,包含认证,授权,访问控制,代理访问等功能"),
		Version:     "1.0.0",
		Description: "To the time to life, rather than to life in time.",
	}}
	s.Definitions =spec.Definitions{"error_message":errorMessageDefinition(),
		userDefinitionModel:userDefinition(),
		userGroupDefinitionModel: userGroupDefinition(),
		policyDefinitionModel:policyDefinition()}
	s.Paths =&spec.Paths{Paths:map[string]spec.PathItem{
		"/auth/register/":{PathItemProps:spec.PathItemProps{Post:NewOperation(
			authTag,
			fmt.Sprintf("认证信息注册"),
			fmt.Sprintf("传入认证标识和密码"),
			[]spec.Parameter{{
			ParamProps: spec.ParamProps{
				In:     "body",
				Name:   "body",
				Description:fmt.Sprintf("参数对象"),
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"object"},
						Ref: getModelSwaggerRef(userDefinitionModel),
					}}}}},
					fmt.Sprintf("无返回"),
					&spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"string"}},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
							})}},
		"/auth/login/":{PathItemProps:spec.PathItemProps{Post:NewOperation(
			authTag,
			fmt.Sprintf("登录认证"),
			fmt.Sprintf("传入认证标识和密码"),
			[]spec.Parameter{{
				ParamProps: spec.ParamProps{
					In:     "body",
					Name:   "body",
					Description:fmt.Sprintf("参数对象"),
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Ref: getModelSwaggerRef(userDefinitionModel),
						}}}}},
						fmt.Sprintf("无返回"),
						&spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"}},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
						})}},
		"/auth/password/":{PathItemProps:spec.PathItemProps{Put:NewOperation(
			authTag,
			fmt.Sprintf("修改密码"),
			fmt.Sprintf("传入认证标识,密码和新密码"),
			[]spec.Parameter{{
				ParamProps: spec.ParamProps{
					In:     "body",
					Name:   "body",
					Description:fmt.Sprintf("参数对象"),
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Ref: getModelSwaggerRef(userDefinitionModel),
						}}}}},
						fmt.Sprintf("无返回"),
						&spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"}},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
						})}},
		"/auth/policy/":{PathItemProps:spec.PathItemProps{Post:NewOperation(
			authTag,
			fmt.Sprintf("检查权限策略"),
			fmt.Sprintf("检查用户或用户组是否具备path的method权限"),
			[]spec.Parameter{{
				ParamProps: spec.ParamProps{
					In:     "body",
					Name:   "body",
					Description:fmt.Sprintf("参数对象"),
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Ref: getModelSwaggerRef(policyDefinitionModel),
						}}}}},
					fmt.Sprintf("无返回"),
					&spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"string"}},
						SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
					})}},
		"/policy/":{PathItemProps:spec.PathItemProps{Get:NewOperation(
			policyTag,
			fmt.Sprintf("获取权限策略"),
			fmt.Sprintf("权限策略用于访问控制"),
			[]spec.Parameter{
				{
					SimpleSchema: spec.SimpleSchema{
						Type: "string",
					},
					ParamProps: spec.ParamProps{
						In:          "query",
						Name:        "user",
						Required:    false,
						Description: "以用户标识筛选",
					}},
				{
					SimpleSchema: spec.SimpleSchema{
						Type: "string",
					},
					ParamProps: spec.ParamProps{
						In:          "query",
						Name:        "path",
						Required:    false,
						Description: "以path筛选",
					}},
			},
			fmt.Sprintf("返回权限策略列表"),
			&spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Ref: getModelSwaggerRef(policyDefinitionModel),
					},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: false},
			}),
			Put:NewOperation(
				policyTag,
				fmt.Sprintf("增加权限策略"),
				fmt.Sprintf("增加用于访问控制的权限策略"),
				[]spec.Parameter{{
					ParamProps: spec.ParamProps{
						In:     "body",
						Name:   "body",
						Description:fmt.Sprintf("参数对象"),
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Ref: getModelSwaggerRef(policyDefinitionModel),
							}}}}},
				fmt.Sprintf("无返回"),
				&spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"string"}},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
				}),
			Delete:NewOperation(
				policyTag,
				fmt.Sprintf("删除权限策略"),
				fmt.Sprintf("删除后使用默认访问控制"),
				[]spec.Parameter{{
					ParamProps: spec.ParamProps{
						In:     "body",
						Name:   "body",
						Description:fmt.Sprintf("参数对象"),
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Ref: getModelSwaggerRef(policyDefinitionModel),
							}}}}},
						fmt.Sprintf("无返回"),
						&spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"}},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
						}),
			}},
		"/policy/group/":{PathItemProps:spec.PathItemProps{Get:NewOperation(
			policyGroupTag,
			fmt.Sprintf("获取组策略"),
			fmt.Sprintf("组策略用于管理用户组"),
			[]spec.Parameter{
				{
					SimpleSchema: spec.SimpleSchema{
						Type: "string",
					},
					ParamProps: spec.ParamProps{
						In:          "query",
						Name:        "user",
						Required:    false,
						Description: "以用户标识筛选",
					}},
				{
					SimpleSchema: spec.SimpleSchema{
						Type: "string",
					},
					ParamProps: spec.ParamProps{
						In:          "query",
						Name:        "group",
						Required:    false,
						Description: "以用户组标识筛选",
					}},
			},
			fmt.Sprintf("返回用户组策略列表"),
			&spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Ref: getModelSwaggerRef(userGroupDefinitionModel),
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: false},
			}),
			Put:NewOperation(
				policyGroupTag,
				fmt.Sprintf("增加组策略"),
				fmt.Sprintf("增加用于管理用户组的组策略"),
				[]spec.Parameter{{
					ParamProps: spec.ParamProps{
						In:     "body",
						Name:   "body",
						Description:fmt.Sprintf("参数对象"),
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Ref: getModelSwaggerRef(userGroupDefinitionModel),
							}}}}},
				fmt.Sprintf("无返回"),
				&spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"string"}},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
				}),
			Delete:NewOperation(
				policyGroupTag,
				fmt.Sprintf("删除组策略"),
				fmt.Sprintf("删除用于管理用户组的组策略"),
				[]spec.Parameter{{
					ParamProps: spec.ParamProps{
						In:     "body",
						Name:   "body",
						Description:fmt.Sprintf("参数对象"),
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Ref: getModelSwaggerRef(userGroupDefinitionModel),
							}}}}},
				fmt.Sprintf("无返回"),
				&spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"string"}},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: ""},
				}),
		}},
		"/policy/metadata/":{PathItemProps:spec.PathItemProps{Head:NewOperation(
				metadataTag,
				fmt.Sprintf("从DB加载最新的元数据"),
				fmt.Sprintf("策略更库后,,如需立即生效,则使用当前api"),
				[]spec.Parameter{},
				fmt.Sprintf("无返回"),
				&spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"integer"},
					},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						Example: 1,
					},
				},
			)}},
	}}
	return
}

func NewOperation(tName,summary, opDescribetion string, params []spec.Parameter,responseDescription string, respSchema *spec.Schema) (op *spec.Operation) {
	op = &spec.Operation{
		spec.VendorExtensible{}, spec.OperationProps{
			Summary:summary,
			Description: opDescribetion,
			//Produces:[]string{"application/json","application/octet-stream"},
			Tags:        []string{tName},
			Parameters:  params,
			Responses: &spec.Responses{
				spec.VendorExtensible{},
				spec.ResponsesProps{
					&spec.Response{
						ResponseProps:spec.ResponseProps{
							Description:"错误消息",
							Schema: &spec.Schema{
								SchemaProps:spec.SchemaProps{
									Ref:getModelSwaggerRef("error_message"),
								},
							},
						},
					},
					map[int]spec.Response{
						200: {
							ResponseProps: spec.ResponseProps{
								Description: responseDescription,
								Schema: respSchema,
							},
						},
						401:{
							ResponseProps: spec.ResponseProps{
								Description: "未认证",
							},
						},
						403:{
							ResponseProps: spec.ResponseProps{
								Description: "未授权",
							},
						},
					},
				},
			},
		},
	}
	return
}

func getModelSwaggerRef(t string) (ref spec.Ref) {
	ref = spec.Ref{}
	ref.Ref, _ = jsonreference.New(fmt.Sprintf("#/definitions/%s", t))
	return
}

func errorMessageDefinition() (schema spec.Schema) {
	//ErrorMessage
	schema.Type = spec.StringOrArray{"object"}
	schema.Title = "错误消息"
	schema.Description = "意外的错误时的消息"
	schema.SchemaProps = spec.SchemaProps{
		Required:[]string{"error"},
		Properties: map[string]spec.Schema{
			"error":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "消息标识",
				},
			},
			"error_description":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "消息描述",
				},
			},
		},
	}
	return
}

func userDefinition() (schema spec.Schema){
	schema.Type = spec.StringOrArray{"object"}
	schema.Title = "认证身份"
	schema.Description = "用于认证的身份信息"
	schema.SchemaProps = spec.SchemaProps{
		Required:[]string{"username","password"},
		Properties: map[string]spec.Schema{
			"username":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "用户名",
				},
			},
			"password":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "密码",
				},
			},
			"new_password":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "新密码",
				},
			},
		},
	}
	return
}

func userGroupDefinition() (schema spec.Schema){
	schema.Type = spec.StringOrArray{"object"}
	schema.Title = "用户和用户组"
	schema.Description = "给用户增加或删除用户组"
	schema.SchemaProps = spec.SchemaProps{
		Required:[]string{},
		Properties: map[string]spec.Schema{
			"user":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "用户标识",
				},
			},
			"group":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "用户组标识",
				},
			},
		},
	}
	return
}

func policyDefinition() (schema spec.Schema){
	schema.Type = spec.StringOrArray{"object"}
	schema.Title = "策略权限"
	schema.Description = "用户或用户组(user)可对某资源路径(path)采取某种操作(method)"
	schema.SchemaProps = spec.SchemaProps{
		Required:[]string{"path","method"},
		Properties: map[string]spec.Schema{
			"user":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "用户或用户组标识",
				},
			},
			"path":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "资源路径",
				},
			},
			"method":spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        spec.StringOrArray{"string"},
					Description: "操作方法",
				},
			},
		},
	}
	return
}

func (s *GatewayServer) getSwaggerJSON() func(c echo.Context) error  {
	return func(c echo.Context) error {
		s.swaggerJSON.Schemes = []string{c.Scheme()}
		return c.JSON(http.StatusOK, s.swaggerJSON)
	}
}
