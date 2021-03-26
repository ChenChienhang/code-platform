// @Author: 陈健航
// @Date: 2021/3/1 0:32
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var CourseResourceController = new(courseResourceController)

type courseResourceController struct{}

func (receiver *courseResourceController) Insert(r *ghttp.Request) {
	var req *model.InsertCourseRecourseReq
	if err := service.CourseResourceService.Insert(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *courseResourceController) Update(r *ghttp.Request) {
	var req *model.UpdateCourseRecourseReq
	if err := service.CourseResourceService.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *courseResourceController) List(r *ghttp.Request) {
	var req *model.ListCourseResourceReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CourseResourceService.List(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *courseResourceController) GetOne(r *ghttp.Request) {
	courseRecourseId := r.GetInt("courseRecourseId")
	resp, err := service.CourseResourceService.GetOne(courseRecourseId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver courseResourceController) Delete(r *ghttp.Request) {
	courseRecourseId := r.GetInt("courseRecourseId")
	if err := service.CourseResourceService.Delete(r.GetCtxVar(dao.SysUser.Columns.UserId).Int(), courseRecourseId); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
