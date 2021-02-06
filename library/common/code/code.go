// @Author: 陈健航
// @Date: 2020/9/27 0:40
// @Description:
package code

import "github.com/gogf/gf/errors/gerror"

var (
	// 10000 是预留的
	OtherError            = gerror.NewCode(10001, "服务器开小差了，请稍后重试")
	VerificationCodeError = gerror.NewCode(10002, "验证码错误")
	UserNotExistError     = gerror.NewCode(10003, "该用户不存在")
	LoginError            = gerror.NewCode(10004, "登陆失败，请稍后重试")
	PermissionError       = gerror.NewCode(10005, "用户权限不足")
	PasswordError         = gerror.NewCode(10006, "密码错误")
	AuthError             = gerror.NewCode(20001, "登录令牌失效，请重新登录")
)
