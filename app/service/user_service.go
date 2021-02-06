// @Author: 陈健航
// @Date: 2020/12/31 0:10
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service/component"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"code-platform/library/common/utils"
	"fmt"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/grand"
	"golang.org/x/crypto/bcrypt"
	"math"
	"mime/multipart"
	"time"
)

type role int

const (
	student role = 1
	teacher      = iota
)

var (
	UserService = new(userService)
	//存放验证码的key头部
	redisVerCodeHeader = "verification_code_"
)

type userService struct{}

// SignUpForStu 注册
// @receiver s
// @params req
// @return error
// @date 2021-01-09 00:15:22
func (s *userService) SignUpForStu(req *model.RegisterReq) error {
	// 校验验证码
	if err := s.checkVerificationCode(req.Email, req.VerificationCode); err != nil {
		return err
	}

	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//转换类型
	insertModel := &model.SysUser{}
	if err = gconv.Struct(req, insertModel); err != nil {
		return err
	}

	//存入加密后的密码
	insertModel.Password = string(hashPassword)
	if _, err = dao.SysUser.OmitEmpty().Insert(insertModel); err != nil {
		return err
	}
	// 把id查出来
	userId, err := dao.SysUser.Where(dao.SysUser.Columns.Email, insertModel.Email).
		FindValue(dao.SysUser.Columns.UserId)
	if err != nil {
		return err
	}
	// 加入casbin权限
	if err = addGroupPolicy(student, userId.Int()); err != nil {
		return err
	}
	return nil
}

// SignUpForStu 教师签发账户注册
// @receiver s
// @params req
// @return error
// @date 2021-01-09 00:15:22
func (s *userService) SignUpForTeacher(req *model.RegisterReq) error {
	// 校验验证码,不用校验验证码

	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//转换类型
	insertModel := &model.SysUser{}
	if err = gconv.Struct(req, insertModel); err != nil {
		return err
	}

	//存入加密后的密码
	insertModel.Password = string(hashPassword)
	if _, err = dao.SysUser.OmitEmpty().Insert(insertModel); err != nil {
		return err
	}
	// 把id查出来
	userId, err := dao.SysUser.
		Where(dao.SysUser.Columns.Email, insertModel.Email).
		FindValue(dao.SysUser.Columns.UserId)
	if err != nil {
		return err
	}
	if err = addGroupPolicy(teacher, userId.Int()); err != nil {
		return err
	}
	return nil
}

// Update 修改个人资料
// @receiver s
// @params req
// @return error
// @date 2021-01-10 00:07:55
func (s *userService) Update(req *model.UserUpdateReq) error {
	// 如果修改了密码，需要进行密码校验
	if req.Password != "" {
		// 在数据库校验用户
		password, err := dao.SysUser.Where(dao.SysUser.Columns.Email).FindValue(dao.SysUser.Columns.Password)
		if err != nil {
			return err
		}
		if password == nil {
			//不存在该用户
			return code.UserNotExistError
		}

		//校验密码 密码错误
		if err = bcrypt.CompareHashAndPassword(password.Bytes(), []byte(req.OldPassword)); err != nil {
			return code.PasswordError
		}
	}
	// 修改了昵称，检查昵称唯一性
	if req.NickName != "" {
		v, err := dao.SysUser.Where(dao.SysUser.Columns.NickName, req.NickName).FindValue(dao.SysUser.Columns.UserId)
		if err != nil {
			return err
		}
		// 如果其他人已经使用了该昵称
		if !v.IsEmpty() || v.Int() != req.UserId {
			return gerror.NewCode(10000, "昵称已存在，请考虑更换其他昵称")
		}
	}
	// 保存
	if _, err := dao.SysUser.OmitEmpty().Save(req); err != nil {
		return err
	}
	return nil
}

// SendVerificationCode 发送注册邮件到邮箱
// @receiver s
// @params emailAddr 邮箱地址
// @return error
// @date 2021-01-09 00:12:55
func (s *userService) SendVerificationCode(emailAddr string) error {
	//获取redis连接
	r := g.Redis()

	//redis键前缀
	redisHeader := redisVerCodeHeader + emailAddr

	//生成6位随机数
	verificationCode := grand.Digits(6)

	// 存入redis，15分钟有效期
	if _, err := r.DoWithTimeout(15*time.Minute, "SET", redisHeader, verificationCode); err != nil {
		return err
	}

	// 准备邮件内容
	emailBody := fmt.Sprintf(response.VerificationCodeEmailBody, verificationCode, gtime.Date())

	// 开一个协程发送邮件
	go func(email string, subject string, emailBody string) {
		if err := utils.SendMail(email, subject, emailBody); err != nil {
			glog.Errorf("发送注册邮件失败:%s", err.Error())
		}
	}(emailAddr, "实验系统邮件验证码", emailBody)

	return nil
}

// ResetPassword 重置密码
// @receiver s
// @params req
// @return error
// @date 2021-01-10 00:07:38
func (s *userService) ResetPassword(req *model.ResetPasswordReq) error {
	// 检查验证码
	if err := s.checkVerificationCode(req.Email, req.VerificationCode); err != nil {
		return err
	}

	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 存入加密后的密码
	if _, err = dao.SysUser.Where(dao.SysUser.Columns.Email, req.Email).
		Save(dao.SysUser.Columns.Password, hashPassword); err != nil {
		return err
	}
	return nil
}

// checkVerificationCode 校验验证码
// @receiver s
// @params emailAddr 邮箱地址
// @params verificationCode 验证码
// @return error
// @date 2021-01-13 20:50:40
func (s *userService) checkVerificationCode(emailAddr string, verificationCode string) error {
	//redis键前缀
	redisHeader := redisVerCodeHeader + emailAddr
	// redis取出验证码
	r := g.Redis()
	v, err := r.DoVar("GET", redisHeader)
	if err != nil {
		return err
	}
	if v.IsEmpty() || v.String() != verificationCode {
		return code.VerificationCodeError
	}
	return nil
}

// DeletedUser 注销用户，形式注销
// @receiver s
// @params req
// @return error
// @date 2021-01-14 11:27:40
func (s *userService) DeletedUser(req *model.DeletedUserReq) error {
	one, err := dao.SysUser.Fields(
		dao.SysUser.Columns.Password,
		dao.SysUser.Columns.Email,
	).WherePri(req.UserId).FindOne()
	if err != nil {
		return err
	}
	// 是否存在该用户
	if one == nil {
		return code.UserNotExistError
	}

	// 校验验证码
	if err = s.checkVerificationCode(one.Email, req.VerificationCode); err != nil {
		return err
	}

	// 校验密码
	if err = bcrypt.CompareHashAndPassword([]byte(one.Password), []byte(req.Password)); err != nil {
		return code.PasswordError
	}

	// 执行删除
	if _, err = dao.SysUser.Delete(req.UserId); err != nil {
		return err
	}
	return nil
}

// ListPage 分页查询所有用户
// @receiver s
// @params current 当前页
// @params size 页面大小
// @return *model.SysUserPageResp
// @return error
// @date 2021-01-15 00:02:22
func (s *userService) ListPage(current, size int) (*model.SysUserPageResp, error) {
	d := dao.SysUser.Page(current, size).
		FieldsEx(dao.SysUser.Columns.Password, dao.LabComment.Columns.DeletedAt)

	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}

	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.SysUserPageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

func (s *userService) ListStuPage(current, size int) (*model.SysUserPageResp, error) {
	d := dao.SysUser.Page(current, size).
		FieldsEx(dao.SysUser.Columns.Password, dao.LabComment.Columns.DeletedAt)

	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}

	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.SysUserPageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

// UploadAvatar 上传头像
// @receiver s
// @params userId 用户id
// @params file 头像文件
// @return error
// @date 2021-02-06 12:09:26
func (s *userService) UploadAvatar(userId int, file multipart.File) error {
	avatarUrl, err := component.MinioUtil.UploadAvatar(file)
	if err != nil {
		return err
	}
	// 把头像保存
	if _, err = dao.SysUser.WherePri(userId).Data(dao.SysUser.Columns.Avatar, avatarUrl).Save(); err != nil {
		return err
	}
	return nil
}

//func (s *userService) isStuInfoCompletable(userId int) (bool, error) {
//
//}

// AddPolicy 添加新用户授权policy
// @receiver s
// @params role 角色，常量
// @params userId
// @date 2021-01-15 00:12:43
func addGroupPolicy(role role, userId int) error {
	if _, err := component.Enforcer.AddGroupingPolicy(userId, role); err != nil {
		return err
	}
	return nil
}
