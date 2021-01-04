package router

import (
	"code-platform/app/api"
	"code-platform/app/service/middleware"
	"code-platform/library/common/component"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

func init() {
	s := g.Server()
	// 启用跨域中间件
	s.Use(middleware.Middleware.CORS)
	s.Group("/user", func(group *ghttp.RouterGroup) {
		component.GfToken.Middleware(group)
		//注册
		group.POST("/signup", api.SysUserController.SignUp)
		//检查昵称唯一性
		group.POST("/check_nickname_unique", api.SysUserController.CheckNickNameUnique)
		//检查邮箱唯一性
		group.POST("/check_email_unique", api.SysUserController.CheckEmailUnique)
	})
}
