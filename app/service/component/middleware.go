// @Author: 陈健航
// @Date: 2020/11/2 11:37
// @Description: 前置中间件
package component

import (
	"github.com/gogf/gf/net/ghttp"
)

// 中间件管理服务
var Middleware = new(serviceMiddleware)

type serviceMiddleware struct{}

// CORS 跨域中间件
// @receiver s
// @params r
// @date 2021-01-04 21:51:50
func (s *serviceMiddleware) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
