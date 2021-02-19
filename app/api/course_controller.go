// @Author: 陈健航
// @Date: 2021/1/12 0:24
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/app/service/component"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var CourseController = new(courseController)

type courseController struct{}

// GetCourse 根据课程id查询相关的课程
// @receiver c courseController
// @params r
// @date 2021-02-08 19:18:26
func (c *courseController) GetCourse(r *ghttp.Request) {
	courseId := r.GetInt("courseId")
	one, err := dao.Course.FieldsEx(dao.Course.Columns.DeletedAt).WherePri(courseId).FindOne()
	if err != nil {
		response.Exit(r, err)
	}
	// 只有开设该课程的教师可以查出密钥
	if one.TeacherId != component.GetUserId(r) {
		one.SecretKey = 0
	}
	response.Success(r, one)
}

// Insert 教师新建课程
// @receiver c
// @params r
// @date 2021-01-14 11:37:04
func (c *courseController) Insert(r *ghttp.Request) {
	// 入参
	var req *model.Course
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = component.GetUserId(r)
	pic := r.GetUploadFile("pic")
	// 保存
	if err := service.CourseService.InsertCourse(req, pic); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// Update 修改课程
// @receiver c
// @params r
// @date 2021-01-14 11:40:27
func (c *courseController) Update(r *ghttp.Request) {
	//入参
	var req *model.Course
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = component.GetUserId(r)
	pic := r.GetUploadFile("pic")
	//保存
	if err := service.CourseService.Update(req, pic); err != nil {
		response.Exit(r, err)
	}

	response.Success(r, true)
}

// ListByTeacherId 根据教师id分页查询所开设课程
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListByTeacherId(r *ghttp.Request) {
	teacherId := component.GetUserId(r)
	current, size := response.GetPageReq(r)
	resp, err := service.CourseService.ListCourseByTeacherId(current, size, teacherId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// ListCourseByStuId 根据学生id分页查询所修读课程
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListCourseByStuId(r *ghttp.Request) {
	stuId := component.GetUserId(r)
	current, size := response.GetPageReq(r)
	resp, err := service.CourseService.ListCourseByStuId(current, size, stuId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// ListStuByCourseId 根据课程id分页获取修读该课程的学生
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListStuByCourseId(r *ghttp.Request) {
	current, size := response.GetPageReq(r)
	courseId := r.GetInt("courseId")
	resp, err := service.CourseService.ListStuPageByCourseId(courseId, current, size)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// StuJoinCourse 学生加入课程
// @receiver c
// @params r
// @date 2021-01-20 22:58:49
func (c *courseController) StuJoinCourse(r *ghttp.Request) {
	// 获取必要信息
	UserId := component.GetUserId(r)
	CourseId := r.GetInt("courseId")
	SecretKey := r.GetInt("secretKey")
	if err := service.CourseService.StuJoinCourse(UserId, CourseId, SecretKey); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// QuitCourse 学生退出课程
// @receiver c courseController
// @params r
// @date 2021-02-06 21:19:55
func (c courseController) QuitCourse(r *ghttp.Request) {
	courseId := r.GetInt("courseId")
	userId := r.GetInt("userId")

	if err := dao.ReCourseUser.DeleteByUserIdAndCourseId(userId, courseId); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// DeleteCourse 教师解散课程，也就是删除课程
// @receiver c labController
// @params r
// @date 2021-01-20 23:07:30
func (c *courseController) DeleteCourse(r *ghttp.Request) {
	courseId := r.GetInt("courseId")
	secretKey := r.GetInt("secretKey")
	teacherId := component.GetUserId(r)
	if err := service.CourseService.DisbandCourseByCourseId(teacherId, courseId, secretKey); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}
