// @Author: 陈健航
// @Date: 2021/2/1 23:44
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var LabController = new(labController)

type labController struct{}

func (receiver *labController) Insert(r *ghttp.Request) {
	var req *model.InsertLabReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.LabService.Insert(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *labController) Update(r *ghttp.Request) {
	var req *model.UpdateLabReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.LabService.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *labController) ListByCourseId(r *ghttp.Request) {
	var req *model.ListLabByCourseIdReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	resp, err := service.LabService.ListByCourseId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *labController) GetOne(r *ghttp.Request) {
	labId := r.GetInt("labId")
	resp, err := service.LabService.GetOne(labId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *labController) Delete(r *ghttp.Request) {
	labId := r.GetInt("labId")
	// 查看开实验的人是不是用户
	if err := service.LabService.Delete(r.GetCtxVar(dao.SysUser.Columns.UserId).Int(), labId); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *labController) ListByToken(r *ghttp.Request) {
	var req *model.ListLabByTokenReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	resp, err := service.LabService.ListByToken(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}
