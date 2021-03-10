package hello

import (
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

type stu struct {
	CourseId int
	response.PageReq
}

// Hello is a demonstration route handler for output "Hello World!".
func Hello(r *ghttp.Request) {
	// 入参
	records := make([]*stu, 0)
	records = append(records, &stu{
		CourseId: 123,
	})
	resp := response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Pages:   5,
			Total:   2,
			Current: 7,
			Size:    78,
		},
	}
	response.Succ(r, resp)

}
