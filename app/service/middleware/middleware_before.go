// @Author: 陈健航
// @Date: 2020/11/2 11:40
// @Description: 后置中间件
package middleware

import (
	"github.com/gogf/gf/net/ghttp"
	"net/http"
)

/**
 * @Description: 统一错误处理
 * @param r
 */
func MiddlewareErrorHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status >= http.StatusInternalServerError {
		r.Response.ClearBuffer()
		r.Response.Writeln("哎哟我去，服务器居然开小差了，请稍后再试吧！")
	}
}
