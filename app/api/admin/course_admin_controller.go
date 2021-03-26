// @Author: 陈健航
// @Date: 2021/3/16 13:45
// @Description:
package admin

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service/admin"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var CourseAdminController = new(courseAdminController)

type courseAdminController struct{}

func (receiver *courseAdminController) ListCourse(r *ghttp.Request) {
	var req *model.ListCourseAdminReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := admin.CourseAdminService.ListCourse(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *courseAdminController) Update(r *ghttp.Request) {
	var req *model.Course
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if _, err := dao.Course.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *courseAdminController) Delete(r *ghttp.Request) {
	courseId := r.GetString("courseId")
	if _, err := dao.Course.WherePri(courseId).Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
