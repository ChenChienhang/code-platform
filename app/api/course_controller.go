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

// InsertCourse 教师新建课程
// @receiver c
// @params r
// @date 2021-01-14 11:37:04
func (c *courseCommentController) InsertCourse(r *ghttp.Request) {
	// 入参
	var req *model.Course
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = component.GetUserId(r)
	// 保存
	if _, err := dao.Course.Insert(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// UpdateCourseByCourseId 修改课程
// @receiver c
// @params r
// @date 2021-01-14 11:40:27
func (c *courseCommentController) UpdateCourseByCourseId(r *ghttp.Request) {
	//入参
	var req *model.Course
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = component.GetUserId(r)
	//保存
	if _, err := dao.Course.Save(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// ListCoursePageByTeacherId 根据教师id分页查询所开设课程
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListCoursePageByTeacherId(r *ghttp.Request) {
	teacherId := component.GetUserId(r)
	current, size := response.GetPageReq(r)
	resp, err := service.CourseService.ListCourseByTeacherId(current, size, teacherId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// ListStuPageByCourseId 根据课程id分页获取修读该课程的学生
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListStuPageByCourseId(r *ghttp.Request) {
	current, size := response.GetPageReq(r)
	courseId := r.GetInt("courseId")
	resp, err := service.CourseService.ListStuPageByCourseId(courseId, current, size)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// JoinCourse 学生加入课程
// @receiver c
// @params r
// @date 2021-01-20 22:58:49
func (c *courseController) JoinCourse(r *ghttp.Request) {
	// 获取必要信息
	UserId := component.GetUserId(r)
	CourseId := r.GetInt("courseId")
	SecretKey := r.GetInt("secretKey")
	if err := service.CourseService.JoinCourse(UserId, CourseId, SecretKey); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// DisbandCourseByCourseId 教师解散课程，也就是删除
// @receiver c labController
// @params r
// @date 2021-01-20 23:07:30
func (c *courseController) DisbandCourseByCourseId(r *ghttp.Request) {
	courseId := r.GetInt("courseId")
	if _, err := dao.Course.WherePri(courseId).And(dao.Course.Columns.TeacherId, component.GetUserId(r)).
		Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// FindStuNotInCourse 找出所有不在该课程的学生
// @receiver c
// @params r
// @date 2021-01-14 15:39:05
//func (c *labController) FindStuNotInCourse(r *ghttp.Request) {
//	var req *model.StuNotInCourseReq
//	if err := r.Parse(&req); err != nil {
//		response.Exit(r, err)
//	}
//	resp, err := service.CourseService.FindStuNotInCourse(req)
//	if err != nil {
//		response.Exit(r, err)
//	}
//	response.Success(r, resp)
//}
