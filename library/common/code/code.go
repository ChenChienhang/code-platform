// @Author: 陈健航
// @Date: 2020/9/27 0:40
// @Description:
package code

import "github.com/gogf/gf/errors/gerror"

var (
	// 10000 是预留的
	OtherError                    = gerror.NewCode(10001, "服务器开小差了，请稍后重试")
	VerificationCodeError         = gerror.NewCode(10002, "验证码错误")
	UserNotExistError             = gerror.NewCode(10003, "该用户不存在")
	LoginError                    = gerror.NewCode(10004, "登陆失败，请稍后重试")
	PermissionError               = gerror.NewCode(10005, "用户权限不足")
	PasswordError                 = gerror.NewCode(10006, "用户密码错误")
	UploadError                   = gerror.NewCode(10007, "上传文件失败，请稍后重试")
	UnSupportUploadError          = gerror.NewCode(10008, "不支持的文件类型")
	InfoNotCompleteError          = gerror.NewCode(10009, "资料未完善，请至少完善学号和真实姓名信息")
	CourseNotExitError            = gerror.NewCode(10010, "课程不存在")
	CourseKeyError                = gerror.NewCode(10011, "加入课程密钥错误")
	NickNameError                 = gerror.NewCode(10012, "昵称已存在，请考虑更换其他昵称")
	UnExpectError                 = gerror.NewCode(10013, "意料外的错误")
	VerificationCodeNotExistError = gerror.NewCode(10014, "验证码错误")
	DDLError                      = gerror.NewCode(10015, "截至时间已过")

	AuthError = gerror.NewCode(20001, "登录令牌失效，请重新登录")
)
