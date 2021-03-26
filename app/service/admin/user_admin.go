// @Author: 陈健航
// @Date: 2021/3/16 12:51
// @Description:
package admin

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"code-platform/library/common/role"
	"github.com/gogf/gf/frame/g"
	"golang.org/x/crypto/bcrypt"
	"math"
)

var UserAdminService = new(userAdminService)

type userAdminService struct{}

func (s *userAdminService) RegisterTeacher(req *model.RegisterReq) (err error) {
	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	//存入加密后的密码
	req.Password = string(hashPassword)
	// 保存
	result, err := dao.SysUser.OmitEmpty().Insert(req)
	if err != nil {
		return err
	}
	// 赋予权限
	teacherId, _ := result.LastInsertId()
	if _, err = g.Table("sys_user_role").Insert(g.Map{
		"user_id": teacherId,
		"role_id": role.Teacher,
	}); err != nil {
		return err
	}
	return nil
}

func (s *userAdminService) RegisterStudent(req *model.RegisterReq) (err error) {
	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	//存入加密后的密码
	req.Password = string(hashPassword)
	// 保存
	result, err := dao.SysUser.OmitEmpty().Insert(req)
	if err != nil {
		return err
	}
	// 赋予权限
	teacherId, _ := result.LastInsertId()
	if _, err = g.Table("sys_user_role").Insert(g.Map{
		"user_id": teacherId,
		"role_id": role.Student,
	}); err != nil {
		return err
	}
	return nil
}

func (s *userAdminService) name() {

}

func (s *userAdminService) ListUser(req *model.ListUserReq) (resp *response.PageResp, err error) {
	userId, err := dao.SysUser.ListUserIdByRole(req.RoleId)
	if err != nil {
		return nil, err
	}
	d := dao.SysUser.WherePri(userId).FieldsEx(dao.SysUser.Columns.Password, dao.LabComment.Columns.DeletedAt)
	records := make([]*model.SysUserResp, 0)
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}

	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}
