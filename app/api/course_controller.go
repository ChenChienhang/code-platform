// @Author: 陈健航
// @Date: 2021/1/12 0:24
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
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
	resp := new(model.CourseResp)
	d := dao.Course.FieldsEx(dao.Course.Columns.DeletedAt).WherePri(courseId)
	// 只有开设该课程的教师可以查出密钥
	if resp.TeacherId != r.GetVar(dao.SysUser.Columns.UserId).Int() {
		d.FieldsEx(dao.Course.Columns.SecretKey)
	}
	if err := d.Scan(&resp); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// Insert 教师新建课程
// @receiver c
// @params r
// @date 2021-01-14 11:37:04
func (c *courseController) Insert(r *ghttp.Request) {
	// 入参
	var req *model.InsertCourseReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	// 保存
	if err := service.CourseService.InsertCourse(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// Update 修改课程
// @receiver c
// @params r
// @date 2021-01-14 11:40:27
func (c *courseController) Update(r *ghttp.Request) {
	//入参
	var req *model.UpdateCourseReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	//保存
	if err := service.CourseService.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// ListByTeacherId 根据教师id分页查询所开设课程
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListByTeacherId(r *ghttp.Request) {
	// 入参
	var req *model.ListByTeacherIdReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	resp, err := service.CourseService.ListCourseByTeacherId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// ListCourseByStuId 根据学生id分页查询所修读课程
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListCourseByStuId(r *ghttp.Request) {
	// 入参
	var req *model.ListCourseByStuIdReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.StudentId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	resp, err := service.CourseService.ListCourseByStuId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// ListStuByCourseId 根据课程id分页获取修读该课程的学生
// @receiver c
// @params r
// @date 2021-01-14 13:36:47
func (c *courseController) ListStuByCourseId(r *ghttp.Request) {
	// 入参
	var req *model.ListStuByCourseIdReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CourseService.ListStuByCourseId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// AttendCourse 学生加入课程
// @receiver c
// @params r
// @date 2021-01-20 22:58:49
func (c *courseController) AttendCourse(r *ghttp.Request) {
	// 入参
	var req *model.AttendCourseReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.StudentId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	if err := service.CourseService.AttendCourse(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// DropCourse 删除选课记录
// @receiver c courseController
// @params r
// @date 2021-02-06 21:19:55
func (c *courseController) DropCourse(r *ghttp.Request) {
	// 入参
	var req *model.DropCourseReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := dao.Course.DeleteByUserIdAndCourseId(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// Delete 教师解散课程，也就是删除课程
// @receiver c labController
// @params r
// @date 2021-01-20 23:07:30
func (c *courseController) Delete(r *ghttp.Request) {
	// 入参
	courseId := r.GetInt("courseId")
	teacherId := r.GetVar(dao.SysUser.Columns.UserId).Int()
	if err := service.CourseService.Delete(teacherId, courseId); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (c *courseController) SearchListByCourseName(r *ghttp.Request) {
	// 入参
	var req *model.SearchListByCourseNameReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CourseService.SearchList(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}
