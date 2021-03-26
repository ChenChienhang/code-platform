// @Author: 陈健航
// @Date: 2020/12/31 0:10
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"code-platform/library/common/role"
	"code-platform/library/common/utils"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/grand"
	"golang.org/x/crypto/bcrypt"
	"math"
	"time"
)

var UserService = newUserService()

type userService struct {
	redisVerCodeHeader string
}

func newUserService() (s *userService) {
	s = &userService{"code.platform:verification.code:"}
	return s
}

// RegisterStudents 注册
// @receiver s
// @params req
// @return error
// @date 2021-01-09 00:15:22
func (s *userService) RegisterStudents(req *model.RegisterReq) (err error) {
	// 校验验证码
	if err = s.checkVerificationCode(req.Email, req.VerificationCode); err != nil {
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
	result, err := dao.SysUser.OmitEmpty().Insert(req)
	if err != nil {
		return err
	}
	// 赋予权限
	stuId, _ := result.LastInsertId()
	if _, err = g.Table("sys_user_role").Insert(g.Map{
		"user_id": stuId,
		"role_id": role.Student,
	}); err != nil {
		return err
	}
	return nil
}

// RegisterStudents 教师签发账户注册
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

	//存入加密后的密码
	req.Password = string(hashPassword)
	if _, err = dao.SysUser.OmitEmpty().Insert(req); err != nil {
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
		//if _, err = dao.SysUser.TX(tx).OmitEmpty().WherePri(req.UserId).Update(req); err != nil {
		//	return err
		//}
		//// 改了真名
		//if req.RealName != nil {
		//	// 课程里的开课老师名字可能要改
		//	if _, err = dao.Course.TX(tx).OmitEmpty().Data(dao.Course.Columns.TeacherName, req.RealName).
		//		Where(dao.Course.Columns.TeacherId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//}
		//// 改了昵称
		//if req.NickName != nil {
		//	// 评论里的名字要改
		//	if _, err = dao.CourseComment.TX(tx).Data(dao.CourseComment.Columns.Username, req.NickName).
		//		Where(dao.CourseComment.Columns.UserId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//	if _, err = dao.CourseComment.TX(tx).Data(dao.CourseComment.Columns.ReplyUsername, req.NickName).
		//		Where(dao.CourseComment.Columns.ReplyId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//	if _, err = dao.LabComment.TX(tx).Data(dao.LabComment.Columns.Username, req.NickName).
		//		Where(dao.LabComment.Columns.UserId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//	if _, err = dao.LabComment.TX(tx).Data(dao.LabComment.Columns.ReplyUsername, req.NickName).
		//		Where(dao.LabComment.Columns.ReplyId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//}
		//// 修改了头像
		//if req.AvatarUrl != nil {
		//	// 评论里的名字要改
		//	if _, err = dao.CourseComment.TX(tx).Data(dao.CourseComment.Columns.UserAvatarUrl, req.AvatarUrl).
		//		Where(dao.CourseComment.Columns.UserId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//	if _, err = dao.LabComment.TX(tx).Data(dao.LabComment.Columns.UserAvatarUrl, req.AvatarUrl).
		//		Where(dao.LabComment.Columns.UserId, req.UserId).Update(); err != nil {
		//		return err
		//	}
		//}
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
	if err := s.SetVerCode(15*time.Minute, emailAddr, verificationCode); err != nil {
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
		Update(dao.SysUser.Columns.Password, hashPassword); err != nil {
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
	v, err := s.GetVerCode(email)
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
func (s *userService) ListUser(current, size int) (*response.PageResp, error) {
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
	resp := &response.PageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

func (s *userService) GetUserInfoByToken(id int) (resp *model.SysUserResp, err error) {
	resp = &model.SysUserResp{}
	if err = dao.SysUser.WherePri(id).FieldsEx(dao.SysUser.Columns.DeletedAt).Scan(&resp); err != nil {
		return nil, err
	}
	if resp.Role, err = dao.SysUser.GetRoleById(id); err != nil {
		return nil, err
	}
	return resp, err
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

func (s *userService) SetVerCode(duration time.Duration, key string, verCode string) (err error) {
	r := g.Redis()
	if duration == -1 {
		if _, err = r.Do("SET", s.redisVerCodeHeader+key, verCode); err != nil {
			return err
		}
	} else {
		if _, err = r.DoWithTimeout(duration, "SET", key, verCode); err != nil {
			return err
		}
	}
	return nil
}

func (s *userService) GetVerCode(key string) (verCode string, err error) {
	r := g.Redis()
	v, err := r.DoVar("GET", s.redisVerCodeHeader+key)
	if err != nil {
		return "", err
	}
	if v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}

func (s *userService) ListCodingTimeByUserId(req *model.ListCodingTimeByUserIdReq) (resp *model.ListCodingTimeByCourseIdResp, err error) {
	resp = &model.ListCodingTimeByCourseIdResp{}
	if err = dao.SysUser.WherePri(req.UserId).Fields(
		dao.SysUser.Columns.UserId,
		dao.SysUser.Columns.Num,
		dao.SysUser.Columns.RealName,
	).FindScan(&resp); err != nil {
		return nil, err
	}
	resp.CodingTime = make([]*model.CodingTimeRecord, 0)

	// 找出所有的编码时间
	codingTime := make([]*model.CodingTime, 0)
	if err = dao.CodingTime.Where(dao.CodingTime.Columns.UserId, req.UserId).
		FieldsEx(dao.CodingTime.Columns.DeletedAt).FindScan(&codingTime); err != nil {
		return nil, err
	}

	// 把编码时间加到每一天上
	for _, v := range codingTime {
		if resp.UserId == v.UserId {
			addFlag := false
			for _, v1 := range resp.CodingTime {
				// 列表中有，累加
				if v1.Date == v.CreatedAt.Format("Y-m-d") {
					v1.Time += v.Duration
					addFlag = true
				}
			}
			// 列表中没有，新增
			if !addFlag {
				resp.CodingTime = append(resp.CodingTime, &model.CodingTimeRecord{
					Date: v.CreatedAt.Format("Y-m-d"),
					Time: v.Duration,
				})
			}
		}
	}
	return resp, nil
}

//func (s *userService) UploadAvatar(avatar *ghttp.UploadFile,userId int) (err error) {
//	uuid, url, err := FileService.UploadPic(avatar, 128)
//	if err != nil {
//		return err
//	}
//	if _, err = dao.MinioFile.Insert(g.Map{
//		dao.MinioFile.Columns.Url:      url,
//		dao.MinioFile.Columns.Filename: avatar.Filename,
//		dao.MinioFile.Columns.Uuid:     uuid,
//	}); err != nil {
//		return err
//	}
//	oldAvatar,err := dao.SysUser.WherePri(userId).FindValue(dao.SysUser.Columns.AvatarUrl)
//	if err != nil {
//		return err
//	}
//	removeFlag :=
//	if oldAvatar != nil{
//		go func() {
//
//		}()
//	}
//	return nil
//}
