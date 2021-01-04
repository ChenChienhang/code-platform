// @Author: 陈健航
// @Date: 2020/11/2 11:37
// @Description: 前置中间件
package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gogf/gf/net/ghttp"
	"net/http"
)

// 中间件管理服务
var Middleware = new(ServiceMiddleware)

type ServiceMiddleware struct {
	enforcer *casbin.Enforcer
}

// CORS 跨域中间件
// @receiver s
// @params r
// @date 2021-01-04 21:51:50
func (s *ServiceMiddleware) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Ctx 上下文中间件
// @receiver s
// @params r
// @date 2021-01-04 21:52:04
func (s *ServiceMiddleware) Ctx(r *ghttp.Request) {
	token := r.Get("token")
	if token == "123456" {
		r.Response.Writeln("auth")
		r.Middleware.Next()
	} else {
		r.Response.WriteStatus(http.StatusForbidden)
	}
}
