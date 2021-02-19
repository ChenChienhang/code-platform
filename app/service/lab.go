// @Author: 陈健航
// @Date: 2021/2/10 22:34
// @Description:
package service

import (
	"code-platform/app/model"
	"mime/multipart"
)

var LabService labService

type labService struct{}

func (s *labService) InsertLab(req *model.Lab, attachment multipart.File) {

}
