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

func (c *theiaController) GetIDEUrl(r *ghttp.Request) {
	var req *model.GetIDEUrlReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	url, err := ide.TheiaService.GetOrRunIDE(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, g.Map{"url": url})
}

func (c *theiaController) CloseIDE(r *ghttp.Request) {
	var req *model.CloseIDEReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	if err := ide.TheiaService.CloseIDE(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
