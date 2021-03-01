// @Author: 陈健航
// @Date: 2021/2/10 23:35
// @Description:
package service

import (
	"bytes"
	"code-platform/app/model"
	"code-platform/app/service/component"
	"code-platform/library/common/code"
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/guuid"
	"github.com/minio/minio-go/v7"
	"image/jpeg"
	"strings"
	"time"
)

var FileService = newFileService()

var (
	picBucketName           = g.Cfg().GetString("minio.bucketName.picBucketName")
	reportBucketName        = g.Cfg().GetString("minio.bucketName.reportBucketName")
	attachmentBucketName    = g.Cfg().GetString("minio.bucketName.attachment")
	videoBucketName         = g.Cfg().GetString("minio.bucketName.videoBucketName")
	redisHeaderDirtyFileSet = "code.platform:dirty.file"
)

type fileService struct{}

func newFileService() (f *fileService) {
	f = new(fileService)
	// 脏文件处理
	go func(redisKey string) {
		for {
			time.Sleep(time.Hour * 24)
			// 获得集合
			v, _ := g.Redis().DoVar("SMEMBERS", redisKey)
			dirtyFiles := make([]model.DirtyFile, 0)
			_ = gconv.Structs(v, &dirtyFiles)
			// 删除所有过期文件
			for _, v := range dirtyFiles {
				// 过去了24小时该文件且未被引用
				if time.Now().Sub(v.CreateTime) > 24*time.Hour {
					// 移除该文件
					go func(service *fileService) {
						_ = f.RemoveObject(v.Url)
					}(f)
					// 从集合中删除
					_, _ = g.Redis().Do("SREM", redisKey, v)
				}
			}
		}
	}(redisHeaderDirtyFileSet)
	return f
}

// RemoveDirtyFile 文件被引用，删除脏文件
// @receiver s
// @date 2021-02-28 16:38:33
func (s *fileService) RemoveDirtyFile(url string) {
	// 获得集合，我觉得这个set应该不会很大，直接遍历问题应该不大
	v, _ := g.Redis().DoVar("SMEMBERS", redisHeaderDirtyFileSet)
	dirtyFiles := make([]model.DirtyFile, 0)
	_ = gconv.Structs(v, &dirtyFiles)
	// 删除所有过期文件
	for _, v := range dirtyFiles {
		// 过去了24小时该文件且未被引用
		if v.Url == url {
			// 从集合中删除
			_, _ = g.Redis().Do("SREM", redisHeaderDirtyFileSet, v)
			break
		}
	}
}

func (s *fileService) AddDirtyFile(url string) {
	// 加入集合
	_, _ = g.Redis().DoVar("SADD", redisHeaderDirtyFileSet, url)
}

// UploadPdf 上传pdf
// @receiver s
// @params uploadFile
// @return url
// @return err
// @date 2021-02-18 23:14:35
func (s *fileService) UploadPdf(uploadFile *ghttp.UploadFile) (url string, err error) {
	if gfile.ExtName(uploadFile.Filename) != "pdf" {
		return "", code.UnSupportUploadError
	}
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
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		reportBucketName,
		uploadName,
		".pdf")
	// 返回可直接访问的url
	go s.AddDirtyFile(url)
	return url, nil
}

// UploadPic 上传图片
// @receiver s
// @params file 文件
// @params width 像素
// @return url
// @return err
// @date 2021-02-18 22:43:25
func (s *fileService) UploadPic(uploadFile *ghttp.UploadFile, width int) (url string, err error) {
	// 文件类型检查
	var contentType string
	switch gfile.ExtName(uploadFile.Filename) {
	case "gif":
		contentType = "image/gif"
	case "png":
		contentType = "image/png"
	case "jpg":
		contentType = "image/jpeg"
	default:
		return "", code.UnSupportUploadError
	}
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
		imageUploadName+gfile.Ext(uploadFile.Filename),
		buff,
		int64(buff.Len()),
		minio.PutObjectOptions{ContentType: contentType},
	); err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		picBucketName,
		imageUploadName,
		gfile.Ext(uploadFile.Filename))
	go s.AddDirtyFile(url)
	// 返回可直接访问的url
	return url, nil
}

// UploadAttachment 上传附件
// @receiver s
// @params uploadFile
// @return url
// @return err
// @date 2021-02-19 11:24:40
func (s *fileService) UploadAttachment(uploadFile *ghttp.UploadFile) (url string, err error) {
	// 文件类型检查,仅支持rar
	if gfile.ExtName(uploadFile.Filename) != "rar" {
		return "", code.UnSupportUploadError
	}
	// 打开，附件一律二进制流格式上传
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
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		picBucketName,
		uploadName,
		gfile.Ext(uploadFile.Filename))
	go s.AddDirtyFile(url)
	// 返回可直接访问的url
	return url, nil
}

func (s *fileService) UploadVideo(uploadFile *ghttp.UploadFile) (url string, err error) {
	// 文件类型检查
	var contentType string
	// 获取后缀名
	switch gfile.ExtName(uploadFile.Filename) {
	case "mp4":
		contentType = "video/mpeg4"
	case "avi":
		contentType = "video/avi"
	default:
		return "", code.UnSupportUploadError
	}
	file, err := uploadFile.Open()
	if err != nil {
		return "", err
	}
	uploadName := strings.ReplaceAll(guuid.New().String(), "-", "")
	if _, err = component.MinioUtil.PutObject(
		context.Background(),
		videoBucketName,
		uploadName+gfile.Ext(uploadFile.Filename),
		file,
		uploadFile.Size,
		minio.PutObjectOptions{ContentType: contentType},
	); err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		picBucketName,
		uploadName,
		gfile.Ext(uploadFile.Filename))
	go s.AddDirtyFile(url)

	return url, nil
}

// RemoveObject 根据url删除文件
// @receiver s
// @params url
// @return error
// @date 2021-02-28 16:32:12
func (s *fileService) RemoveObject(url string) error {
	objectName := s.getObjectName(url)
	var bucketName string
	switch gfile.ExtName(objectName) {
	case "gif":
		bucketName = picBucketName
	case "png":
		bucketName = picBucketName
	case "jpg":
		bucketName = picBucketName
	case "mp4":
		bucketName = videoBucketName
	case "avi":
		bucketName = videoBucketName
	case "pdf":
		bucketName = reportBucketName
	case "rar":
		bucketName = attachmentBucketName
	default:
		// 不支持的文件格式
		return code.UnSupportUploadError
	}
	if err := component.MinioUtil.RemoveObject(
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
