// @Author: 陈健航
// @Date: 2021/2/26 19:42
// @Description:
package api

import (
	"code-platform/app/service"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

var FileController = new(fileController)

type fileController struct{}

func (receiver *fileController) UploadPic(r *ghttp.Request) {
	uploadFile := r.GetUploadFile("pic")
	width := r.GetInt("width")
	_, url, err := service.FileService.UploadPic(uploadFile, width)
	if err != nil {
		response.Exit(r, code.UploadError)
	}
	response.Succ(r, g.Map{"url": url})
}

func (receiver *fileController) UploadAttachment(r *ghttp.Request) {
	uploadFile := r.GetUploadFile("attachment")
	url, err := service.FileService.UploadAttachment(uploadFile)
	if err != nil {
		response.Exit(r, code.UploadError)
	}
	response.Succ(r, g.Map{"url": url})
}

func (receiver *fileController) UploadPdf(r *ghttp.Request) {
	uploadFile := r.GetUploadFile("pdf")
	url, err := service.FileService.UploadPdf(uploadFile)
	if err != nil {
		response.Exit(r, code.UploadError)
	}
	response.Succ(r, g.Map{"url": url})
}

func (receiver *fileController) UploadVideo(r *ghttp.Request) {
	uploadFile := r.GetUploadFile("video")
	url, err := service.FileService.UploadVideo(uploadFile)
	if err != nil {
		response.Exit(r, code.UploadError)
	}
	response.Succ(r, g.Map{"url": url})
}

func (receiver *fileController) DeleteObject(r *ghttp.Request) {
	url := r.GetString("url")
	if err := service.FileService.RemoveObject(url); err != nil {
		response.Exit(r, code.UploadError)
	}
	response.Succ(r, true)
}
