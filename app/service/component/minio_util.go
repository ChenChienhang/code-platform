// @Author: 陈健航
// @Date: 2021/1/12 21:34
// @Description:
package component

import (
	"context"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

//goland:noinspection ALL
const (
	avatarBucketName    = "pic"
	reportBucketName    = "report"
	attchmentBucketName = "attchment"
	// 自己试出来的，官网的sdk说明已经过时了，不能直接把"github.com/minio/minio-go/v7/pkg/policy"的常量放进去
	read_only      = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	write_only     = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	write_and_read = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\",\"s3:GetBucketLocation\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
)

var MinioUtil = initMinioUtil()

func initMinioUtil() (m *minio.Client) {
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

	location := "cn-north-1"
	ctx := context.Background()

	// 创建一个存储桶,用于放图片
	bucketName := avatarBucketName
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
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(read_only, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
	// 创建一个叫report的存储桶,用于放实验报告
	bucketName = reportBucketName
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
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(read_only, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
	return m
}

//
//// UploadPic 上传头像，返回url
//// @receiver u
//// @params img
//// @params width
//// @return url 头像url
//// @return err
//// @date 2021-01-13 18:19:48
//func (u *minioUtil) UploadPic(file multipart.File, width int) (url string, err error) {
//	// 编码
//	img, err := jpeg.Decode(file)
//	if err != nil {
//		return "", err
//	}
//	imageUploadName := strings.ReplaceAll(guuid.New().String(), "-", "")
//	// 压缩成缩略图
//	dstImage128 := imaging.Resize(img, width, 0, imaging.Lanczos)
//	buff := new(bytes.Buffer)
//	// 装入缓存
//	if err = jpeg.Encode(buff, dstImage128, nil); err != nil {
//		return "", err
//	}
//	// 上传
//	if _, err = u.minioClient.PutObject(
//		context.Background(),
//		avatarBucketName,
//		imageUploadName+".jpeg",
//		buff,
//		int64(buff.Len()),
//		minio.PutObjectOptions{ContentType: "image/jpeg"},
//	); err != nil {
//		return "", err
//	}
//	// 返回可直接访问的头像url"http://118.178.253.239:9000/avatar/789.jpg"
//	return fmt.Sprintf("%s/%s/%s%s",
//			g.Cfg().GetString("minio.endpoint"),
//			avatarBucketName,
//			imageUploadName,
//			".jpg"),
//		nil
//}
//
//func (u *minioUtil) UploadAttachment(file multipart.File) {
//
//}
//
//func (u *minioUtil) DeletePic(url string) (err error) {
//	objectName := getObjectName(url)
//	if err = u.minioClient.RemoveObject(
//		context.Background(),
//		avatarBucketName,
//		objectName,
//		minio.RemoveObjectOptions{}); err != nil {
//		return
//	}
//	return
//}
//
//// getObjectName
//// @receiver u
//// @params s
//// @return objectName
//// @date 2021-02-09 21:26:56
//func getObjectName(s string) (objectName string) {
//	left := gstr.PosR(s, "/")
//	objectName = gstr.SubStr(s, left+1)
//	return
//}
