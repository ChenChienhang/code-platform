package router

import (
	"code-platform/app/api"
	"code-platform/app/api/hello"
	"code-platform/app/service/component"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

func init() {
	s := g.Server()
	// 启用跨域中间件
	s.Use(component.Middleware.CORS)
	// 前台
	s.Group("/web", func(group *ghttp.RouterGroup) {
		component.GfToken.Middleware(group)
		// 个人用户相关
		group.Group("/user", func(group *ghttp.RouterGroup) {
			// 学生用户注册
			group.POST("/signup/stu", api.SysUserController.StuSignUp)
			// 上传头像
			group.POST("/avatar", api.SysUserController.UpdateAvatarByUserId)
			// 根据用户id获取信息
			group.GET("/{userId}", api.SysUserController.GetOneById)
			// 分页获取所有用户信息
			group.GET("/list", api.SysUserController.ListUserPage)
			// 更新用户信息
			group.PUT("/", api.SysUserController.UpdateById)
			// 注销用户
			group.DELETE("/", api.SysUserController.DeleteById)
			// 检查昵称唯一性
			group.GET("/nickname/{nickname}", api.SysUserController.IsNicknameAccessible)
			// 检查邮箱唯一性
			group.GET("/email/{email}", api.SysUserController.IsEmailAccessible)
			// 获取验证码
			group.POST("/verificationCode", api.SysUserController.SendVerificationCode)
			// 重置密码
			group.PUT("/password", api.SysUserController.ResetPassword)
			// 注销账户
			group.GET("/test", hello.Hello)
		})
		// 评论相关
		s.Group("/comment", func(group *ghttp.RouterGroup) {

		})
	})
}
