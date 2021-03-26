// @Author: 陈健航
// @Date: 2021/3/12 9:45
// @Description:
package admin

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service/admin"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var SysUserAdminController = new(sysUserAdminController)

type sysUserAdminController struct{}

func (receiver *sysUserAdminController) registerTeacher(r *ghttp.Request) {
	var req *model.RegisterReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := admin.UserAdminService.RegisterTeacher(req); err != nil {
		response.Exit(r, err)
	}
}

func (receiver *sysUserAdminController) registerStudent(r *ghttp.Request) {
	var req *model.RegisterReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := admin.UserAdminService.RegisterTeacher(req); err != nil {
		response.Exit(r, err)
	}
}

func (receiver *sysUserAdminController) ListUser(r *ghttp.Request) {
	var req *model.ListUserReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := admin.UserAdminService.ListUser(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *sysUserAdminController) Update(r *ghttp.Request) {
	var req *model.SysUser
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if _, err := dao.SysUser.OmitEmpty().Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *sysUserAdminController) Delete(r *ghttp.Request) {
	userId := r.GetString("userId")
	if _, err := dao.SysUser.WherePri(userId).Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
