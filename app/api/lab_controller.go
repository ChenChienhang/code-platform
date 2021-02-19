// @Author: 陈健航
// @Date: 2021/2/1 23:44
// @Description:
package api

import (
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
	"mime/multipart"
)

var LabController = new(labController)

type labController struct{}

func (c *labController) InsertLab(r *ghttp.Request) {
	var req *model.Lab
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	file := r.GetUploadFile("attachment")
	var attachment multipart.File = nil
	if file != nil {
		var err error
		attachment, err = file.Open()
		if err != nil {
			response.Exit(r, err)
		}
	}
	service.LabService.InsertLab(req, attachment)

}
