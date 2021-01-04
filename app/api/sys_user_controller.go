// @Author: 陈健航
// @Date: 2020/12/30 23:29
// @Description:
package api

import (
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/component"
	"code-platform/library/common/response"
	"fmt"
	"github.com/gogf/gf/net/ghttp"
)

var SysUserController = new(apiSysUserController)

type apiSysUserController struct{}

// @summary 注册
// @description 提交注册信息，注册用户。
// @tags    用户相关
// @Accept  json
// @produce json
// @param   Email		   	formData 	string 	true  "邮箱"
// @param   NickName   		formData  	string 	true  "昵称"
// @param   RealName		formData  	string 	false "真实姓名"
// @param   Num				formData 	string 	true  "学号/职工号"
// @param   Password		formData  	string 	true  "密码"
// @param   Major query 	formData	string	true  "专业"
// @param   Organization 	formData	string 	true  "单位"
// @router  /user/signup [POST]
// @success 200 {object} response.JsonResponse "执行结果"
func (s *apiSysUserController) SignUp(r *ghttp.Request) {
	//入参
	var sysUser *model.SysUserRegisterReq
	_ = r.Parse(&sysUser)
	// 服务层处理
	err := service.UserService.SignUp(sysUser)
	if err != nil {
		fmt.Print(err)
		response.ExitSpec(r, err.Error())
	}
	response.Success(r)
}

// @summary 昵称唯一性检查
// @description 检查用户昵称的唯一性。
// @tags    用户相关
// @Accept  json
// @produce json
// @param   nick_name formData string true "用户昵称"
// @router  /user/check_nickname_unique [POST]
// @success 200 {object} bool "执行结果"
func (s *apiSysUserController) CheckNickNameUnique(r *ghttp.Request) {
	// 入参
	var NickName *model.SysUserCheckNickNameReq
	_ = r.Parse(&NickName)
	res, err := service.UserService.CheckNickUnique(NickName)
	if err != nil {
		fmt.Print(err)
	}
	response.Success(r, res)
}

// @summary 邮箱唯一性检查
// @description 检查用户邮箱的唯一性。
// @tags    用户相关
// @Accept  json
// @produce json
// @param   nick_name formData string true "用户邮箱"
// @router  /user/check_email_unique [POST]
// @success 200 {object} bool "执行结果"
func (s *apiSysUserController) CheckEmailUnique(r *ghttp.Request) {
	component.GfToken.GetTokenData(r)
	// 入参
	var Email *model.SysUserCheckEmailReq
	_ = r.Parse(&Email)
	res, err := service.UserService.CheckEmailUnique(Email)
	if err != nil {
		fmt.Print(err)
	}
	response.Success(r, res)
}

// @summary 注册
// @description 注意保存的数据通过表单提交，由于提交的数据字段不固定，因此这里没有写字段说明，并且无法通过`swagger`测试。
// @tags    注册
// @produce json
// @param   table    path  string true  "操作的数据表"
// @param   x_schema query string false "操作的数据库"
// @router  /curd/{table}/save [POST]
// @success 200 {object} response.JsonResponse "执行结果"
//func Modify(r *ghttp.Request) {
//	var sysUser *model.UserModifyReq
//	_ = r.Parse(&sysUser)
//	err := service.UserService.SignUp(sysUser)
//	if err != nil {
//		fmt.Print(err)
//		response.ExitSpec(r, err.Error())
//	}
//	response.Success(r)
//}
