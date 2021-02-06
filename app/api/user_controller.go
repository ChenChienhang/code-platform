// @Author: 陈健航
// @Date: 2020/12/30 23:29
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

var SysUserController = new(sysUserController)

type sysUserController struct{}

// StuSignUp 学生身份注册
// @receiver c
// @params r
// @date 2021-01-14 00:06:43
func (c *sysUserController) StuSignUp(r *ghttp.Request) {
	//入参
	var sysUser *model.RegisterReq
	if err := r.Parse(&sysUser); err != nil {
		response.Exit(r, err)
	}
	// 服务层处理
	if err := service.UserService.SignUpForStu(sysUser); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// GetOneById 根据id查询用户信息
// @receiver c
// @params r
// @date 2021-01-14 00:10:04
func (c *sysUserController) GetOneById(r *ghttp.Request) {
	userId := r.GetInt("userId")
	// 不传默认从token取
	if userId == 0 {
		userId = component.GetUserId(r)
	}
	// 字段脱敏
	one, err := dao.SysUser.FieldsEx(
		dao.SysUser.Columns.DeletedAt,
		dao.SysUser.Columns.Password,
	).WherePri(userId).FindOne()
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, one)
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
	response.Success(r, v.IsEmpty())
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
	response.Success(r, v.IsEmpty())
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
	response.Success(r, true)
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
	response.Success(r, true)
}

// UpdateById 修改信息
// @params r
// @date 2021-01-14 00:07:57
func (c *sysUserController) UpdateById(r *ghttp.Request) {
	var req *model.UserUpdateReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.UserId = component.GetUserId(r)

	if err := service.UserService.Update(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
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
	req.UserId = component.GetUserId(r)

	if err := service.UserService.DeletedUser(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// ListUserPage 分页查询所有用户
// @receiver c
// @params r
// @date 2021-01-14 12:48:12
func (c *sysUserController) ListUserPage(r *ghttp.Request) {
	current, size := response.GetPageReq(r)
	resp, err := service.UserService.ListPage(current, size)

	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// ListUserPage 分页查询学生用户
// @receiver c
// @params r
// @date 2021-01-14 12:48:12
func (c *sysUserController) ListStuPage(r *ghttp.Request) {
	current, size := response.GetPageReq(r)
	resp, err := service.UserService.ListStuPage(current, size)

	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// UpdateAvatarByUserId 更新头像
// @receiver c
// @params r
// @date 2021-02-06 12:13:58
func (c *sysUserController) UpdateAvatarByUserId(r *ghttp.Request) {
	avatar := r.GetUploadFile("avatar")
	avatarFile, err := avatar.Open()
	if err != nil {
		response.Exit(r, err)
	}
	defer func() {
		_ = avatarFile.Close()
	}()
	userId := component.GetUserId(r)
	if err = service.UserService.UploadAvatar(userId, avatarFile); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}
