// @Author: 陈健航
// @Date: 2021/3/5 20:07
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service/ide"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

var TheiaController = new(theiaController)

type theiaController struct{}

func (receiver *theiaController) OpenIDE(r *ghttp.Request) {
	var req *model.OpenIDEReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	url, err := ide.TheiaService.OpenIDE(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, g.Map{"url": url})
}

func (receiver *theiaController) CloseIDE(r *ghttp.Request) {
	var req *model.CloseIDEReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	if role, err := dao.SysUser.GetRoleById(req.UserId); err != nil {
		response.Exit(r, err)
	} else if role == 2 {
		// 学生才经常stop，教师的等程序清理stop
		if err = ide.TheiaService.StopIDE(req); err != nil {
			response.Exit(r, err)
		}
	}
	response.Succ(r, true)
}

func (receiver *theiaController) CheckCode(r *ghttp.Request) {
	var req *model.CheckCodeReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.TeacherId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	url, err := ide.TheiaService.CheckCode(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, g.Map{"url": url})
}
