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
			// 用户注册，学生，已添加
			group.POST("/signup/stu", api.SysUserController.StuSignUp)
			// 根据用户id获取信息，所有，已添加
			group.GET("/{userId}", api.SysUserController.GetOne)
			// 分页获取所有用户信息，未确定
			//group.GET("/all", api.SysUserController.ListUser)
			// 更新用户信息，所有
			group.PUT("/", api.SysUserController.UpdateById)
			// 注销用户,所有
			group.DELETE("/", api.SysUserController.DeleteById)
			// 检查昵称唯一性，无
			group.GET("/nickname", api.SysUserController.IsNicknameAccessible)
			// 检查邮箱唯一性，无
			group.GET("/email", api.SysUserController.IsEmailAccessible)
			// 获取验证码，无
			group.POST("/verificationCode", api.SysUserController.SendVerificationCode)
			// 重置密码，无
			group.PUT("/password", api.SysUserController.ResetPassword)
			// 测试
			group.POST("/test", hello.Hello)
		})
		// 评论相关
		group.Group("/comment", func(group *ghttp.RouterGroup) {
			// 新增课程评论
			group.POST("/course", api.CommentController.InsertCourseComment)
			// 新增实验评论
			group.POST("/lab", api.CommentController.InsertLabComment)
			// 查询实验评论
			group.GET("/course", api.CommentController.ListCourseComment)
			group.GET("/lab", api.CommentController.ListLabComment)
			group.DELETE("/course", api.CommentController.DeleteCourseComment)
			group.DELETE("/lab", api.CommentController.DeleteLabComment)
		})
		// 课程相关
		group.Group("/course", func(group *ghttp.RouterGroup) {
			// 根据课程id查询课程信息,教师，学生,已添加
			group.GET("/{courseId}", api.CourseController.GetCourse)
			// 新建课程，教师,已添加
			group.POST("/", api.CourseController.Insert)
			// 修改课程，教师,已添加
			group.PUT("/", api.CourseController.Update)
			// 教师解散课程，也就是删除课程，教师
			group.DELETE("/", api.CourseController.Delete)
			// 根据教师id分页查询所有该老师开设课程，教师,已添加
			group.GET("/setup", api.CourseController.ListByTeacherId)
			// 根据用户id分页查询所该用户修读课程，学生,已添加
			group.GET("/study", api.CourseController.ListCourseByStuId)
			// 根据课程id分页获取修读该课程的学生，教师,已添加
			group.GET("/student/:courseId", api.CourseController.ListStuByCourseId)
			// 学生加入课程，学生,已添加
			group.POST("/attend", api.CourseController.AttendCourse)
			// 学生退出课程，学生,已添加
			group.DELETE("/quit", api.CourseController.DropCourse)

		})
		// 签到相关
		group.Group("/checkin", func(group *ghttp.RouterGroup) {
			// 学生签到
			group.GET("/check", api.CheckInController.CheckinForStudent)
			// 教师发起签到
			group.POST("/start", api.CheckInController.CheckinForTeacher)
			// 签到记录
			group.GET("/records", api.CheckInController.ListCheckinRecords)
			// 签到详情
			group.GET("/details", api.CheckInController.ListCheckinDetails)
			// 删除签到记录
			group.DELETE("/record", api.CheckInController.DeleteRecordsDetails)
			// 修改签到详情
			group.PUT("/detail", api.CheckInController.UpdateCheckinDetail)
			// 导出签到表
			group.GET("/export/{courseId}", api.CheckInController.ExportCheckinRecords)
		})
		// 实验相关
		group.Group("/lab", func(group *ghttp.RouterGroup) {
			// 新增实验
			group.POST("/", api.LabController.Insert)
			// 修改实验
			group.PUT("/", api.LabController.Update)
			// 列表
			group.GET("/", api.LabController.List)
			// 获取实验详细信息
			group.GET("/{labId}", api.LabController.GetOne)
			// 删除实验
			group.DELETE("/{labId}", api.LabController.Delete)
		})
		group.Group("/upload", func(group *ghttp.RouterGroup) {
			// 上传图片
			group.POST("/pic", api.FileController.UploadPic)
			// 上传pdf
			group.POST("/pdf", api.FileController.UploadPdf)
			// 上传附件
			group.POST("/attachments", api.FileController.UploadAttachment)
			// 上传视频
			group.POST("/video", api.FileController.UploadVideo)
			group.DELETE("/", api.FileController.DeleteObject)
		})
		group.Group("/ide", func(group *ghttp.RouterGroup) {
			// 打开ide
			group.POST("/", api.TheiaController.GetIDEUrl)
			// 关闭ide
			group.DELETE("/", api.TheiaController.CloseIDE)
		})
		group.Group("/summit", func(group *ghttp.RouterGroup) {
			// 上传报告
			group.POST("/report", api.LabSummitController.InsertReport)
			// 代码提交（其实就是做个标记
			group.POST("/code", api.LabSummitController.InsertCodeFinish)
			// 查阅完成情况
			group.GET("/", api.LabSummitController.SelectLabSummit)
			// 收集编译错误信息
			group.GET("/compile/log", api.LabSummitController.SelectCompilerErrorLog)
		})
	})

}
