// @Author: 陈健航
// @Date: 2021/2/10 23:35
// @Description:
package service

import (
	"bytes"
	"code-platform/app/service/component"
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/guuid"
	"github.com/minio/minio-go/v7"
	"image/jpeg"
	"strings"
)

var FileService = new(fileService)

var (
	picBucketName        = "pic"
	reportBucketName     = "report"
	attachmentBucketName = "attachment"
)

type fileService struct{}

// UploadPdf 上传pdf
// @receiver s
// @params uploadFile
// @return url
// @return err
// @date 2021-02-18 23:14:35
func (s *fileService) UploadPdf(uploadFile *ghttp.UploadFile) (url string, err error) {
	// 打开
	file, err := uploadFile.Open()
	if err != nil {
		return "", err
	}
	// 文件名
	uploadName := strings.ReplaceAll(guuid.New().String(), "-", "")
	// 上传
	if _, err = component.MinioUtil.PutObject(
		context.Background(),
		reportBucketName,
		fmt.Sprintf("%s%s", uploadName, ".pdf"),
		file,
		uploadFile.Size,
		minio.PutObjectOptions{ContentType: "application/pdf"},
	); err != nil {
		return "", err
	}
	// 返回可直接访问的url
	return fmt.Sprintf("%s/%s/%s%s",
			g.Cfg().GetString("minio.endpoint"),
			reportBucketName,
			uploadName,
			".pdf"),
		nil
}

// UploadPic 上传图片
// @receiver s
// @params file 文件
// @params width 像素
// @return url
// @return err
// @date 2021-02-18 22:43:25
func (s *fileService) UploadPic(uploadFile *ghttp.UploadFile, width int) (url string, err error) {
	// 编码
	file, err := uploadFile.Open()
	if err != nil {
		return "", err
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		return "", err
	}
	// 文件名
	imageUploadName := strings.ReplaceAll(guuid.New().String(), "-", "")
	// 压缩成缩略图
	dstImage128 := imaging.Resize(img, width, 0, imaging.Lanczos)
	buff := new(bytes.Buffer)
	// 装入缓存
	if err = jpeg.Encode(buff, dstImage128, nil); err != nil {
		return "", err
	}
	// 上传
	if _, err = component.MinioUtil.PutObject(
		context.Background(),
		picBucketName,
		imageUploadName+".jpeg",
		buff,
		int64(buff.Len()),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	); err != nil {
		return "", err
	}
	// 返回可直接访问的url
	return fmt.Sprintf("%s/%s/%s%s",
			g.Cfg().GetString("minio.endpoint"),
			picBucketName,
			imageUploadName,
			".jpeg"),
		nil
}

// UploadAttachment 上传附件
// @receiver s
// @params uploadFile
// @return url
// @return err
// @date 2021-02-19 11:24:40
func (s *fileService) UploadAttachment(uploadFile *ghttp.UploadFile) (url string, err error) {
	// 打开
	file, err := uploadFile.Open()
	if err != nil {
		return "", err
	}
	// 文件名
	uploadName := strings.ReplaceAll(guuid.New().String(), "-", "")
	// 上传
	if _, err = component.MinioUtil.PutObject(
		context.Background(),
		picBucketName,
		// gfile.Ext获得后缀名，包括.
		uploadName+gfile.Ext(uploadFile.Filename),
		file,
		uploadFile.Size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	); err != nil {
		return "", err
	}
	// 返回可直接访问的url
	return fmt.Sprintf("%s/%s/%s%s",
			g.Cfg().GetString("minio.endpoint"),
			picBucketName,
			uploadName,
			gfile.Ext(uploadFile.Filename)),
		nil
}

//RemovePic 移除图片
//@receiver s
//@params url
//@return error
//@date 2021-02-10 23:42:03
func (s *fileService) RemovePic(url string) (err error) {
	objectName := s.getObjectName(url)
	if err = s.RemoveObject(objectName, picBucketName); err != nil {
		return err
	}
	return nil
}

// RemovePdf 移除pdf
// @receiver s
// @params url
// @return err
// @date 2021-02-19 11:25:51
func (s *fileService) RemovePdf(url string) (err error) {
	objectName := s.getObjectName(url)
	if err = s.RemoveObject(objectName, reportBucketName); err != nil {
		return err
	}
	return nil
}

// RemoveAttachment 移除attachment
// @receiver s
// @params url
// @return err
// @date 2021-02-19 11:25:51
func (s *fileService) RemoveAttachment(url string) (err error) {
	objectName := s.getObjectName(url)
	if err = s.RemoveObject(objectName, attachmentBucketName); err != nil {
		return err
	}
	return nil
}

// RemoveObject 删除
// @receiver s
// @params objectName 文件名
// @params bucketName 存储桶
// @return err
// @date 2021-02-19 11:23:00
func (s *fileService) RemoveObject(objectName string, bucketName string) (err error) {
	if err = component.MinioUtil.RemoveObject(
		context.Background(),
		bucketName,
		objectName,
		minio.RemoveObjectOptions{}); err != nil {
		return err
	}
	return nil
}

// getObjectName 根据url获取文件名
// @receiver u
// @params s
// @return objectName
// @date 2021-02-09 21:26:56
func (s *fileService) getObjectName(name string) (objectName string) {
	left := gstr.PosR(name, "/")
	objectName = gstr.SubStr(name, left+1)
	return objectName
}
