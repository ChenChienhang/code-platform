// @Author: 陈健航
// @Date: 2020/12/29 11:00
// @Description:
package test

import (
	"code-platform/library/common/utils"
	"fmt"
	"testing"
)

type JsonResponse struct {
	Code    int         `json:"code"`    // 错误码((0:成功, 1:失败, >1:错误码))
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 返回数据(业务接口定义具体数据结构)
}

func Test123(t *testing.T) {
	err := utils.SendMail("853804445@qq.com", "123", "123")
	fmt.Print(err)
}
