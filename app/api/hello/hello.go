package hello

import (
	"code-platform/library/common/response"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gogf/gf/net/ghttp"
)

type stu struct {
	CourseId *int `v:"required"`
}

// Hello is a demonstration route handler for output "Hello World!".
func Hello(r *ghttp.Request) {
	// 入参
	file := excelize.NewFile()
	// 准备表头
	if err := file.SetCellStr("Sheet1", "A1", "学号"); err != nil {
		response.Exit(r, err)
	}
	//buffer, err := file.WriteToBuffer()
	//if err != nil {
	//	response.Exit(r,err)
	//}
	//if err = r.Write(buffer); err != nil {
	//	response.Exit(r,err)
	//}
	if err := file.Write(r.Response.Writer); err != nil {
		response.Exit(r, err)
	}
	r.Response.Header().Set("Content-disposition", "attachment;filename="+"签到表.xlsx")
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Exit()
}
