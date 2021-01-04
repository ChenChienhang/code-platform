// @Author: 陈健航
// @Date: 2020/9/27 0:40
// @Description:
package response

type ErrorCode struct {
	// 错误消息
	ErrorMessage string
	// 错误码
	ErrorCode int
}

var (
	PasswordError     = &ErrorCode{"密码错误", 10002}
	UserNotExistError = &ErrorCode{"该用户不存在", 10003}
	LoginError        = &ErrorCode{"登陆失败，请稍后重试", 10004}
	AuthError         = &ErrorCode{"登录令牌失效，请重新登录", 20001}
	PermissionError   = &ErrorCode{"用户权限不足", 20001}
)
