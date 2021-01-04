// @Author: 陈健航
// @Date: 2020/11/1 22:55
// @Description:
package response

import (
	"github.com/gogf/gf/net/ghttp"
)

/**
 * @Description: 数据返回通用JSON数据结构
 */
type JsonResponse struct {
	Code    int         `json:"code"`    // 错误码(0成功，其他错误)
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 返回数据(业务接口定义具体数据结构)
}

// Success 成功返回结果集
// @params r
// @params data
// @date 2021-01-04 22:16:50
func Success(r *ghttp.Request, data ...interface{}) {
	// data为空时返回空集合,go里面入参interface后要通过反射判空
	var responseData = interface{}(nil)
	if len(data) > 0 {
		responseData = data
	} else {
		responseData = make([]interface{}, 0)
	}
	// 写回json
	_ = r.Response.WriteJson(
		JsonResponse{
			Code:    000000,
			Message: "执行成功",
			Data:    responseData,
		},
	)
}

// Exit 发生错误返回结果集
// @params r
// @params error
// @date 2021-01-04 22:17:08
func Exit(r *ghttp.Request, error *ErrorCode) {
	// 封装错误信息
	_ = r.Response.WriteJson(
		JsonResponse{
			Code:    error.ErrorCode,
			Message: error.ErrorMessage,
			Data:    make([]interface{}, 0),
		},
	)
	// 返回给前端
	r.Exit()
}

// ExitSpec 具体输入返回结果集
// @params r
// @params errMes
// @date 2021-01-04 22:17:34
func ExitSpec(r *ghttp.Request, errMes string) {
	// 封装错误信息
	_ = r.Response.WriteJson(
		JsonResponse{
			Code:    10001,
			Message: errMes,
			Data:    make([]interface{}, 0),
		},
	)
	// 返回给前端
	r.Exit()
}
