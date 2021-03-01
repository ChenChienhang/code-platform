// @Author: 陈健航
// @Date: 2020/12/30 23:29
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var SysUserController = new(sysUserController)

type sysUserController struct{}

// StuSignUp 学生身份注册
// @receiver c
// @params r
// @date 2021-01-14 00:06:43
func (c *sysUserController) StuSignUp(r *ghttp.Request) {
	//入参
	var req *model.SignUpReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	// 服务层处理
	if err := service.UserService.SignUpForStu(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// GetOne 根据id查询用户信息
// @receiver c
// @params r
// @date 2021-01-14 00:10:04
func (c *sysUserController) GetOne(r *ghttp.Request) {
	userId := r.GetInt("userId")
	// 不传默认从token取
	if userId == 0 {
		userId = r.GetVar(dao.SysUser.Columns.UserId).Int()
	}
	// 字段脱敏
	resp := new(model.SysUserResp)
	if err := dao.SysUser.FieldsEx(
		dao.SysUser.Columns.DeletedAt,
		dao.SysUser.Columns.Password,
	).WherePri(userId).Scan(&resp); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// IsNicknameAccessible 昵称唯一性检查
// @receiver c
// @params r
// @date 2021-01-14 00:07:22
func (c *sysUserController) IsNicknameAccessible(r *ghttp.Request) {
	nickName := r.GetString("nickname")
	v, err := dao.SysUser.Where(dao.SysUser.Columns.NickName, nickName).FindValue(dao.SysUser.Columns.UserId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, v.IsEmpty())
}

// isEmailAccessible 邮箱唯一性检查
// @receiver c
// @params r
// @date 2021-01-14 00:07:30
func (c *sysUserController) IsEmailAccessible(r *ghttp.Request) {
	email := r.GetString("email")
	v, err := dao.SysUser.Where(dao.SysUser.Columns.Email, email).FindValue(dao.SysUser.Columns.UserId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, v.IsEmpty())
}

// SendVerificationCode 获取验证码
// @receiver c
// @params r
// @date 2021-01-14 00:07:38
func (c *sysUserController) SendVerificationCode(r *ghttp.Request) {
	emailAddr := r.GetString("email")
	if err := service.UserService.SendVerificationCode(emailAddr); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// ResetPassword 重置密码，在之前需要进行邮箱验证码发送
// @receiver c
// @params r
// @date 2021-01-14 00:32:33
func (c *sysUserController) ResetPassword(r *ghttp.Request) {
	var req *model.ResetPasswordReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}

	if err := service.UserService.ResetPassword(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// UpdateById 修改信息
// @params r
// @date 2021-01-14 00:07:57
func (c *sysUserController) UpdateById(r *ghttp.Request) {
	var req *model.UserUpdateReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetVar(dao.SysUser.Columns.UserId).Int()

	if err := service.UserService.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// DeleteById 注销用户
// @receiver c
// @params r
// @date 2021-01-14 11:36:06
func (c *sysUserController) DeleteById(r *ghttp.Request) {
	var req *model.DeletedUserReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = r.GetVar(dao.SysUser.Columns.UserId).Int()

	if err := service.UserService.DeletedUser(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// ListUser 分页查询所有用户
// @receiver c
// @params r
// @date 2021-01-14 12:48:12
func (c *sysUserController) ListUser(r *ghttp.Request) {
	//current, size := response.GetPageReq(r)
	//resp, err := service.UserService.ListUser(current, size)
	//
	//if err != nil {
	//	response.Exit(r, err)
	//}
	//response.Succ(r, resp)
}
