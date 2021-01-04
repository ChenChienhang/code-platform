// @Author: 陈健航
// @Date: 2020/12/31 23:14
// @Description:
package component

import (
	"code-platform/app/service"
	"code-platform/library/common/response"
	"fmt"
	"github.com/goflyfox/gtoken/gtoken"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
)

var GfToken = &gtoken.GfToken{
	LoginPath:        "/login",
	LogoutPath:       "/logout",
	LoginBeforeFunc:  service.UserService.LoginBeforeFunc,
	LoginAfterFunc:   LoginAfterFunc,
	LogoutAfterFunc:  LogoutAfterFunc,
	AuthAfterFunc:    AuthAfterFunc,
	GlobalMiddleware: true,
}

// LoginAfterFunc 重定义返回后结果集
// @params r
// @params respData
// @date 2021-01-04 22:14:51
func LoginAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	if !respData.Success() {
		response.Exit(r, response.LoginError)
	} else {
		// 返回token
		response.Success(r, g.Map{
			"token": respData.GetString("token"),
		})
	}
}

// LogoutAfterFunc 重定义退登结果集
// @params r
// @params respData
// @date 2021-01-04 22:14:32
func LogoutAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	if respData.Success() {
		response.Success(r)
	} else {
		response.Exit(r, response.LoginError)
	}
}

// AuthAfterFunc 身份认证操作的后续,主要是鉴权
// @params r
// @params respData
// @date 2021-01-04 22:13:39
func AuthAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	//存在令牌
	if respData.Success() {
		// 鉴权
		AuthRes, err := authenticate(respData.GetString("userKey"), r.URL.Path, r.Method)
		if err != nil {
			response.ExitSpec(r, err.Error())
		}
		if !AuthRes {
			// 权限不足
			response.Exit(r, response.PermissionError)
		}
		r.Middleware.Next()
		//不存在令牌
	} else {
		var params map[string]interface{}
		if r.Method == "GET" {
			params = r.GetMap()
		} else if r.Method == "POST" {
			params = r.GetMap()
		} else {
			response.ExitSpec(r, "Request Method is ERROR! ")
			return
		}
		no := gconv.String(gtime.TimestampMilli())
		glog.Info(fmt.Sprintf("[AUTH_%s][url:%s][params:%s][data:%s]",
			no, r.URL.Path, params, respData.Json()))
		response.Exit(r, response.AuthError)
	}
}

// authenticate 鉴权方法
// @params userId
// @params url
// @params method
// @return bool
// @return error
// @date 2021-01-04 21:58:46
func authenticate(userId string, url string, method string) (bool, error) {
	res, err := Enforcer.Enforce(userId, url, method)
	if err != nil {
		return false, err
	}
	return res, nil
}
