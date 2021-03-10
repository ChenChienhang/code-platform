// @Author: 陈健航
// @Date: 2021/2/10 23:35
// @Description:
package service

import (
	"bytes"
	"code-platform/app/model"
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
	"github.com/minio/minio-go/v7/pkg/credentials"
	"image/jpeg"
	"strings"
	"time"
)

var FileService = newFileService()

type fileService struct {
	minio                   *minio.Client
	picBucketName           string
	reportBucketName        string
	attachmentBucketName    string
	videoBucketName         string
	redisHeaderDirtyFileSet string
	policyReadOnly          string
	policyWriteOnly         string
	policyWriteAndRead      string
}

func newFileService() (f *fileService) {

	endpoint := g.Cfg().GetString("minio.endpoint")
	accessKeyID := g.Cfg().GetString("minio.accessKeyID")
	secretAccessKey := g.Cfg().GetString("minio.secretAccessKey")
	// 初使化 minio client对象。
	m, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	f = &fileService{
		minio:                   m,
		picBucketName:           g.Cfg().GetString("minio.bucketName.picBucketName"),
		reportBucketName:        g.Cfg().GetString("minio.bucketName.reportBucketName"),
		attachmentBucketName:    g.Cfg().GetString("minio.bucketName.attachment"),
		videoBucketName:         g.Cfg().GetString("minio.bucketName.videoBucketName"),
		redisHeaderDirtyFileSet: "code.platform:dirty.file",
		// 自己试出来的，官网的sdk说明已经过时了，不能直接把"github.com/minio/minio-go/v7/pkg/policy"的常量放进去
		policyReadOnly:     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}",
		policyWriteOnly:    "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}",
		policyWriteAndRead: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\",\"s3:GetBucketLocation\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}",
	}

	location := "cn-north-1"
	ctx := context.Background()

	// 创建一个存储桶,用于放图片
	bucketName := f.picBucketName
	isExist, err := m.BucketExists(ctx, bucketName)
	if err != nil {
		panic(err)
	}
	if !isExist {
		if err = m.MakeBucket(
			ctx,
			bucketName,
			minio.MakeBucketOptions{Region: location},
		); err != nil {
			panic(bucketName + " 存储桶 创建失败" + err.Error())
		}
		//设置该存储桶策略
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(f.policyReadOnly, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
	// 创建一个叫report的存储桶,用于放实验报告
	bucketName = f.reportBucketName
	isExist, err = m.BucketExists(ctx, bucketName)
	if err != nil {
		panic(err)
	}
	if !isExist {
		if err = m.MakeBucket(
			ctx,
			bucketName,
			minio.MakeBucketOptions{Region: location},
		); err != nil {
			panic(bucketName + " 存储桶 创建失败" + err.Error())
		}
		//设置该存储桶策略
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(f.policyReadOnly, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}

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
	}(f.redisHeaderDirtyFileSet)
	return f
}

// RemoveDirtyFile 文件被引用，删除脏文件
// @receiver s
// @date 2021-02-28 16:38:33
func (s *fileService) RemoveDirtyFile(url string) {
	// 获得集合，我觉得这个set应该不会很大，直接遍历问题应该不大
	v, _ := g.Redis().DoVar("SMEMBERS", s.redisHeaderDirtyFileSet)
	dirtyFiles := make([]model.DirtyFile, 0)
	_ = gconv.Structs(v, &dirtyFiles)
	// 删除所有过期文件
	for _, v := range dirtyFiles {
		// 过去了24小时该文件且未被引用
		if v.Url == url {
			// 从集合中删除
			_, _ = g.Redis().Do("SREM", s.redisHeaderDirtyFileSet, v)
			break
		}
	}
}

func (s *fileService) AddDirtyFile(url string) {
	// 加入集合
	_, _ = g.Redis().DoVar("SADD", s.redisHeaderDirtyFileSet, url)
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
	if _, err = s.minio.PutObject(
		context.Background(),
		s.reportBucketName,
		fmt.Sprintf("%s%s", uploadName, ".pdf"),
		file,
		uploadFile.Size,
		minio.PutObjectOptions{ContentType: "application/pdf"},
	); err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		s.reportBucketName,
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
	if _, err = s.minio.PutObject(
		context.Background(),
		s.picBucketName,
		imageUploadName+gfile.Ext(uploadFile.Filename),
		buff,
		int64(buff.Len()),
		minio.PutObjectOptions{ContentType: contentType},
	); err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		s.picBucketName,
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
	if _, err = s.minio.PutObject(
		context.Background(),
		s.picBucketName,
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
		s.picBucketName,
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
	if _, err = s.minio.PutObject(
		context.Background(),
		s.videoBucketName,
		uploadName+gfile.Ext(uploadFile.Filename),
		file,
		uploadFile.Size,
		minio.PutObjectOptions{ContentType: contentType},
	); err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		s.videoBucketName,
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
		bucketName = s.picBucketName
	case "png":
		bucketName = s.picBucketName
	case "jpg":
		bucketName = s.picBucketName
	case "mp4":
		bucketName = s.videoBucketName
	case "avi":
		bucketName = s.videoBucketName
	case "pdf":
		bucketName = s.reportBucketName
	case "rar":
		bucketName = s.attachmentBucketName
	default:
		// 不支持的文件格式
		return code.UnSupportUploadError
	}
	if err := s.minio.RemoveObject(
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
