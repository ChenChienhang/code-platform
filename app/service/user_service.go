// @Author: 陈健航
// @Date: 2020/12/31 0:10
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"fmt"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"golang.org/x/crypto/bcrypt"
)

var UserService = new(userService)

type userService struct{}

func (s *userService) SignUp(req *model.SysUserRegisterReq) error {
	// 密码加盐加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	//存入加密后的密码
	req.Password = string(hash)
	_, err = dao.SysUser.Save(req)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	return nil
}

// CheckNickUnique 邮箱是否唯一
// @receiver s
// @params req
// @return bool
// @return error
// @date 2021-01-04 22:25:52
func (s *userService) CheckNickUnique(req *model.SysUserCheckNickNameReq) (bool, error) {
	res, err := dao.SysUser.Fields(dao.SysUser.Columns.UserId).FindOne(dao.SysUser.Columns.NickName, req.Nickname)
	if err != nil {
		return false, err
	}
	return res == nil, nil
}

func (s *userService) CheckEmailUnique(req *model.SysUserCheckEmailReq) (bool, error) {
	res, err := dao.SysUser.Fields(dao.SysUser.Columns.UserId).FindOne(dao.SysUser.Columns.Email, req.Email)
	if err != nil {
		return false, err
	}
	return res == nil, nil
}

// @summary 登录
// @description 登录，返回token。
// @tags    快速CURD
// @Accept  json
// @produce json
// @param   username    formData  	 string true  "用户名（邮箱）"
// @param   password 	formData     string false "密码"
// @router  /user/login [POST]
// @success 200 {object} string "token"
func (s *userService) LoginBeforeFunc(r *ghttp.Request) (string, interface{}) {
	var req *model.LoginReq
	// 转换成结构体
	if err := r.Parse(&req); err != nil {
		response.ExitSpec(r, err.Error())
	}
	// 在数据库校验用户
	res, err := dao.SysUser.FindOne(dao.SysUser.Columns.Email, req.Username)
	if err != nil {
		response.ExitSpec(r, err.Error())
	}
	if res == nil {
		//不存在该用户
		response.Exit(r, response.UserNotExistError)
	}
	// 校验密码 密码错误
	if //goland:noinspection GoNilness
	err = bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(req.Password)); err != nil {
		response.Exit(r, response.PasswordError)
	}
	//校验成功
	return gconv.String(res.UserId), ""
}

func (s *userService) Modify(req *model.UserModifyReq) error {
	// 如果修改了新密码
	if req.NewPassword != "" {
		// 在数据库校验用户
		res, err := dao.SysUser.FindOne(dao.SysUser.Columns.Email)
		if err != nil {
			return err
		}
		if res == nil {
			//不存在该用户
			return err
		}
		// 校验密码 密码错误
		//if //goland:noinspection GoNilness
		//err = bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(req.Password)); err != nil {
		//	response.Exit(r, error_code.PasswordError)
		//}
	}
	return nil
}
