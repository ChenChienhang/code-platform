// @Author: 陈健航
// @Date: 2020/11/1 22:55
// @Description:
package response

import (
	code2 "code-platform/library/common/code"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
)

// jsonResponse 数据返回通用JSON数据结构
type jsonResponse struct {
	Code    int         `json:"code"`    // 错误码(0成功，其他错误)
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 返回数据(业务接口定义具体数据结构)
}

// PageInfo 分页参数
type PageInfo struct {
	Total   int `json:"total"`   // 总记录数
	Size    int `json:"size"`    // 当前页面条数
	Current int `json:"current"` // 当前页码
	Pages   int `json:"pages"`   // 总共页码
}

// Success 成功返回结果集
// @params r
// @params data
// @date 2021-01-04 22:16:50
func Success(r *ghttp.Request, data ...interface{}) {
	var responseData = interface{}(nil)
	if len(data) > 0 {
		responseData = data[0]
	} else {
		responseData = make([]interface{}, 0)
	}
	// 写回json
	_ = r.Response.WriteJson(
		jsonResponse{
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
func Exit(r *ghttp.Request, err error) {
	//打印错误日志
	glog.Errorf("[url:%s][err:%s]",
		r.URL.Path, err.Error())
	if code := gerror.Code(err); code == -1 {
		err = code2.OtherError
	}
	// 封装错误信息,返回给前端
	_ = r.Response.WriteJson(
		jsonResponse{
			Code:    gerror.Code(err),
			Message: err.Error(),
			Data:    false,
		},
	)
}

func GetPageReq(r *ghttp.Request) (int, int) {
	current := r.GetInt("pageCurrent")
	if current <= 0 {
		current = 1
	}
	size := r.GetInt("pageSize")
	if size <= 0 {
		size = 10
	}
	return current, size
}
