// @Author: 陈健航
// @Date: 2021/3/8 16:39
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/app/service/ide"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var LabSummitController = new(labSummitController)

type labSummitController struct{}

func (c *labSummitController) InsertReport(r *ghttp.Request) {
	var req *model.SummitReportReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.StuId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	req.Report = r.GetUploadFile("report")
	if err := service.LabSummitService.InsertReport(req); err != nil {
		response.Exit(r, err)
	}
}

func (c *labSummitController) InsertCodeFinish(r *ghttp.Request) {
	var req *model.SummitLabFinishReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.StuId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	if err := service.LabSummitService.InsertCodeFinish(req); err != nil {
		response.Exit(r, err)
	}
}

func (c *labSummitController) SelectLabSummit(r *ghttp.Request) {
	var req *model.SelectLabSummitReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.LabSummitService.SelectLabSummit(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (c *labSummitController) SelectCompilerErrorLog(r *ghttp.Request) {
	var req *model.SelectCompilerErrorLogReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := ide.TheiaService.CollectCompilerErrorLog(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}
