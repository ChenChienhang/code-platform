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
		//component.GfToken.Middleware(group)
		// 个人用户相关
		group.Group("/user", func(group *ghttp.RouterGroup) {
			// 用户注册，学生，已添加
			group.POST("/signup/stu", api.SysUserController.StuSignUp)
			// 根据用户id获取信息，所有，已添加
			group.GET("/:userId", api.SysUserController.GetOneById)
			// 分页获取所有用户信息，未确定
			group.GET("/list", api.SysUserController.ListUser)
			// 更新用户信息，所有
			group.PUT("/", api.SysUserController.UpdateById)
			// 注销用户,所有
			group.DELETE("/", api.SysUserController.DeleteById)
			// 检查昵称唯一性，无
			group.GET("/nickname/:nickname", api.SysUserController.IsNicknameAccessible)
			// 检查邮箱唯一性，无
			group.GET("/email/:email", api.SysUserController.IsEmailAccessible)
			// 获取验证码，无
			group.POST("/verificationCode", api.SysUserController.SendVerificationCode)
			// 重置密码，无
			group.PUT("/password", api.SysUserController.ResetPassword)
			// 测试
			group.POST("/test", hello.Hello)
		})
		// 评论相关
		group.Group("/comment", func(group *ghttp.RouterGroup) {

		})
		group.Group("/course", func(group *ghttp.RouterGroup) {
			// 根据课程id查询课程信息,教师，学生,已添加
			group.GET("/:courseId", api.CourseController.GetCourse)
			// 新建课程，教师,已添加
			group.POST("/", api.CourseController.Insert)
			// 修改课程，教师,已添加
			group.PUT("/", api.CourseController.Update)
			// 根据教师id分页查询所有该老师开设课程，教师,已添加
			group.GET("/setup", api.CourseController.ListByTeacherId)
			// 根据用户id分页查询所该用户修读课程，学生,已添加
			group.GET("/study", api.CourseController.ListCourseByStuId)
			// 根据课程id分页获取修读该课程的学生，教师,已添加
			group.GET("/student/:courseId", api.CourseController.ListStuByCourseId)
			// 学生加入课程，学生,已添加
			group.POST("/join", api.CourseController.StuJoinCourse)
			// 学生退出课程，学生,已添加
			group.DELETE("/quit/:courseId", api.CourseController.QuitCourse)
			// 教师解散课程，也就是删除课程，教师
			group.DELETE("/", api.CourseController.DeleteCourse)
		})
	})
}
