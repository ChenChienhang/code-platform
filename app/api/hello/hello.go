package hello

import (
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

// Hello is a demonstration route handler for output "Hello World!".
func Hello(r *ghttp.Request) {
	//file := r.GetUploadFile("we")
	id, err := service.UserService.ListPage(1, 2)
	if err != nil {
		return
	}
	response.Success(r, id)
}
