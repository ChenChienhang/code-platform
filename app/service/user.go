// @Author: 陈健航
// @Date: 2020/12/31 0:10
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"code-platform/library/common/utils"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/grand"
	"golang.org/x/crypto/bcrypt"
	"math"
	"time"
)

type role int

const (
	Teacher role = 1 + iota
	Student
)

var UserService = newUserService()

type userService struct {
	IUserServiceCache
}

func newUserService() (s *userService) {
	s = &userService{newIUserServiceCache()}
	return s
}

// SignUpForStu 注册
// @receiver s
// @params req
// @return error
// @date 2021-01-09 00:15:22
func (s *userService) SignUpForStu(req *model.SignUpReq) error {
	// 校验验证码
	if err := s.checkVerificationCode(req.Email, req.VerificationCode); err != nil {
		return err
	}

	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	//存入加密后的密码
	req.Password = string(hashPassword)
	// 保存
	if _, err = dao.SysUser.OmitEmpty().Insert(req); err != nil {
		return err
	}
	// 把id查出来
	//if userId, err := dao.SysUser.Where(dao.SysUser.Columns.Email, req.Email).
	//	FindValue(dao.SysUser.Columns.UserId); err != nil {
	//	return err
	//} else {
	// 加入casbin权限
	//if err = addGroupPolicy(Student, userId.Int()); err != nil {
	//	return err
	//}
	//}
	return nil
}

// SignUpForStu 教师签发账户注册
// @receiver s
// @params req
// @return error
// @date 2021-01-09 00:15:22
func (s *userService) SignUpForTeacher(req *model.SignUpReq) error {
	// 校验验证码,不用校验验证码

	// 密码加密
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//存入加密后的密码
	req.Password = string(hashPassword)
	if _, err = dao.SysUser.OmitEmpty().Insert(req); err != nil {
		return err
	}
	// 把id查出来
	//if userId, err := dao.SysUser.
	//	Where(dao.SysUser.Columns.Email, req.Email).
	//	FindValue(dao.SysUser.Columns.UserId); err != nil {
	//	return err
	//} else {
	//	if err = addGroupPolicy(Teacher, userId.Int()); err != nil {
	//		return err
	//	}
	//}
	return nil
}

// Update 修改个人资料
// @receiver s
// @params req
// @return error
// @date 2021-01-10 00:07:55
func (s *userService) Update(req *model.UserUpdateReq) error {
	// 如果修改了密码，需要进行密码校验
	if req.Password != nil {
		// 在数据库校验用户
		password, err := dao.SysUser.WherePri(req.UserId).FindValue(dao.SysUser.Columns.Password)
		if err != nil {
			return err
		}
		if password == nil {
			//不存在该用户
			return code.UserNotExistError
		}

		//校验密码 密码错误
		if err = bcrypt.CompareHashAndPassword(password.Bytes(), []byte(*req.OldPassword)); err != nil {
			return code.PasswordError
		}
	}
	// 修改了昵称，检查昵称唯一性
	if req.NickName != nil {
		v, err := dao.SysUser.Where(dao.SysUser.Columns.NickName, req.NickName).FindValue(dao.SysUser.Columns.UserId)
		if err != nil {
			return err
		}
		// 如果其他人已经使用了该昵称
		if !v.IsEmpty() || v.Int() != req.UserId {
			return code.NickNameError
		}
	}
	// 修改了头像
	removeFlag := false
	if req.AvatarUrl != nil {
		// 保存旧头像url
		if oldPicUrl, err := dao.SysUser.WherePri(req.UserId).FindValue(dao.SysUser.Columns.AvatarUrl); err != nil {
			return err
		} else if !oldPicUrl.IsEmpty() {
			removeFlag = true
			defer func(flag *bool) {
				if *flag {
					go FileService.RemoveDirtyFile(*req.AvatarUrl)
					//goland:noinspection GoUnhandledErrorResult
					go FileService.RemoveObject(oldPicUrl.String())
				}
			}(&removeFlag)
		}
	}
	// 保存
	if tx, err := g.DB().Begin(); err != nil {
		removeFlag = false
		return err
	} else {
		if _, err = dao.SysUser.TX(tx).OmitEmpty().Save(req); err != nil {
			return err
		}
		// 改了真名
		if req.RealName != nil {
			// 课程里的开课老师名字可能要改
			if _, err = dao.Course.TX(tx).OmitEmpty().Data(dao.Course.Columns.TeacherName, req.RealName).
				Where(dao.Course.Columns.TeacherId, req.UserId).Save(); err != nil {
				return err
			}
		}
		// 改了昵称
		if req.NickName != nil {
			// 评论里的名字要改
			if _, err = dao.CourseComment.TX(tx).OmitEmpty().Data(dao.CourseComment.Columns.Username, req.NickName).
				Where(dao.CourseComment.Columns.UserId, req.UserId).Save(); err != nil {
				return err
			}
			if _, err = dao.CourseComment.TX(tx).OmitEmpty().Data(dao.CourseComment.Columns.ReplyUsername, req.NickName).
				Where(dao.CourseComment.Columns.ReplyId, req.UserId).Save(); err != nil {
				return err
			}
			if _, err = dao.LabComment.TX(tx).OmitEmpty().Data(dao.LabComment.Columns.Username, req.NickName).
				Where(dao.LabComment.Columns.UserId, req.UserId).Save(); err != nil {
				return err
			}
			if _, err = dao.LabComment.TX(tx).OmitEmpty().Data(dao.LabComment.Columns.ReplyUsername, req.NickName).
				Where(dao.LabComment.Columns.ReplyId, req.UserId).Save(); err != nil {
				return err
			}
		}
		// 修改了头像
		if req.AvatarUrl != nil {
			if _, err = dao.CourseComment.TX(tx).OmitEmpty().Data(dao.CourseComment.Columns.UserAvatarUrl, req.AvatarUrl).
				Where(dao.CourseComment.Columns.UserId, req.UserId).Save(); err != nil {
				return err
			}
			if _, err = dao.LabComment.TX(tx).OmitEmpty().Data(dao.LabComment.Columns.UserAvatarUrl, req.AvatarUrl).
				Where(dao.LabComment.Columns.UserId, req.UserId).Save(); err != nil {
				return err
			}
		}
		// 提交
		if err = tx.Commit(); err != nil {
			removeFlag = false
			return err
		}
	}
	return nil
}

// SendVerificationCode 发送注册邮件到邮箱
// @receiver s
// @params emailAddr 邮箱地址
// @return error
// @date 2021-01-09 00:12:55
func (s *userService) SendVerificationCode(emailAddr string) error {

	//生成6位随机数
	verificationCode := grand.Digits(6)

	// 存入redis，15分钟有效期
	if err := s.IUserServiceCache.SetV(15*time.Minute, emailAddr, verificationCode); err != nil {
		return err
	}
	// 准备邮件内容
	emailBody := fmt.Sprintf(`<div style="background-color:#ECECEC; padding: 35px;">
    <table cellpadding="0" align="center"
           style="width: 600px; margin: 0px auto; text-align: left; position: relative; border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 5px; border-bottom-left-radius: 5px; font-size: 14px; font-family:微软雅黑, 黑体; line-height: 1.5; box-shadow: rgb(153, 153, 153) 0px 0px 5px; border-collapse: collapse; background-position: initial initial; background-repeat: initial initial;background:#fff;">
        <tbody>
        <tr>
            <th valign="middle"
                style="height: 25px; line-height: 25px; padding: 15px 35px; border-bottom-width: 1px; border-bottom-style: solid; border-bottom-color: #42a3d3; background-color: #49bcff; border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 0px; border-bottom-left-radius: 0px;">
                <font face="微软雅黑" size="5" style="color: rgb(255, 255, 255); ">注册成功! （阿里云）</font>
            </th>
        </tr>
        <tr>
            <td>
                <div style="padding:25px 35px 40px; background-color:#fff;">
                    <h2 style="margin: 5px 0px; ">
                        <font color="#333333" style="line-height: 20px; ">
                            <font style="line-height: 22px; " size="4">
                                亲爱的 123456</font>
                        </font>
                    </h2>
                    <p>首先感谢您使用本站！这是您的验证码：%s<br>
                    <p align="right">%s</p>
                    <div style="width:700px;margin:0 auto;">
                        <div style="padding:10px 10px 0;border-top:1px solid #ccc;color:#747474;margin-bottom:20px;line-height:1.3em;font-size:12px;">
                            <p>此为系统邮件，请勿回复<br>
                                请保管好您的邮箱，避免账号被他人盗用
                            </p>
                            <p>©***</p>
                        </div>
                    </div>
                </div>
            </td>
        </tr>
        </tbody>
    </table>
</div>
`, verificationCode, gtime.Datetime())
	// 开一个协程发送邮件
	go func() {
		_ = utils.MailUtil.SendMail(emailAddr, "实验系统邮件验证码", emailBody)
	}()
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
func (s *userService) checkVerificationCode(email string, verificationCode string) error {
	v, err := s.GetV(email)
	if err != nil {
		return nil
	}
	if v == "" {
		return code.VerificationCodeNotExistError
	}
	if v != verificationCode {
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

// ListUser 分页查询所有用户
// @receiver s
// @params current 当前页
// @params size 页面大小
// @return *model.SysUserPageResp
// @return error
// @date 2021-01-15 00:02:22
func (s *userService) ListUser(current, size int) (*model.SysUserPageResp, error) {
	d := dao.SysUser.Page(current, size).
		FieldsEx(dao.SysUser.Columns.Password, dao.LabComment.Columns.DeletedAt)

	all := make([]*model.SysUserResp, 0)
	if err := d.Scan(&all); err != nil {
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

//func (s *userService) ListStuPage(current, size int) (*model.SysUserPageResp, error) {
//	d := dao.SysUser.Page(current, size).
//		FieldsEx(dao.SysUser.Columns.Password, dao.LabComment.Columns.DeletedAt)
//
//	all, err := d.FindAll()
//	if err != nil {
//		return nil, err
//	}
//	if all == nil {
//		all = make([]*model.SysUser, 0)
//	}
//
//	count, err := d.Count()
//	if err != nil {
//		return nil, err
//	}
//
//	// 分页信息整合
//	resp := &model.SysUserPageResp{
//		Records: all,
//		PageInfo: &response.PageInfo{
//			Size:    len(all),
//			Total:   count,
//			Current: current,
//			Pages:   int(math.Ceil(float64(count) / float64(size))),
//		}}
//	return resp, nil
//}

type IUserServiceCache interface {
	SetV(duration time.Duration, key string, value interface{}) (err error)
	GetV(key string) (value interface{}, err error)
}

func newIUserServiceCache() (c IUserServiceCache) {
	if g.Cfg().GetBool("server.Multiple") {
		c = &userServiceRedisCache{
			key: "code.platform:verification.code:",
		}
		return c
	} else {
		c = &userServiceSimpleCache{
			gcache.New(),
		}
		return c
	}
}

type userServiceRedisCache struct {
	key string
}

func (u *userServiceRedisCache) SetV(duration time.Duration, key string, value interface{}) (err error) {
	r := g.Redis()
	if duration == -1 {
		if _, err = r.Do("SET", key, value); err != nil {
			return err
		}
	} else {
		if _, err = r.DoWithTimeout(duration, "SET", key, value); err != nil {
			return err
		}
	}

	return nil
}

func (u *userServiceRedisCache) GetV(key string) (value interface{}, err error) {
	r := g.Redis()
	v, err := r.DoVar("GET", key)
	if err != nil {
		return "", err
	}
	if v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}

type userServiceSimpleCache struct {
	*gcache.Cache
}

func (u *userServiceSimpleCache) SetV(duration time.Duration, key string, value interface{}) (err error) {
	if duration == -1 {
		if err = u.Set(0, value, duration); err != nil {
			return err
		}
	} else {
		if err = u.Set(key, value, duration); err != nil {
			return err
		}
	}
	return nil
}

func (u *userServiceSimpleCache) GetV(key string) (value interface{}, err error) {
	v, err := u.GetVar(key)
	if err != nil {
		return "", err
	}
	if v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}
