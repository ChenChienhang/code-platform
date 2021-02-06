// @Author: 陈健航
// @Date: 2021/1/12 21:34
// @Description:
package component

import (
	"bytes"
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/guuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"image/jpeg"
	"mime/multipart"
	"os"
	"strings"
)

// 自己试出来的，官网的sdk说明已经过时了，不能直接把"github.com/minio/minio-go/v7/pkg/policy"的常量放进去
//goland:noinspection ALL
const (
	avatarBucketName = "avatar"
	reportBucketName = "report"
	read_only        = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	write_only       = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	write_and_read   = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\",\"s3:GetBucketLocation\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
)

var MinioUtil *minioUtil

type minioUtil struct {
	minioClient *minio.Client
}

func InitMinioUtil() {
	endpoint := g.Cfg().GetString("minio.endpoint")
	accessKeyID := g.Cfg().GetString("minio.accessKeyID")
	secretAccessKey := g.Cfg().GetString("minio.secretAccessKey")
	// 初使化 minio client对象。
	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	MinioUtil = &minioUtil{minioClient: c}

	// 创建一个叫avatar的存储桶,用于放头像
	bucketName, location, ctx := "avatar", "cn-north-1", context.Background()
	isExist, err := MinioUtil.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		panic(err)
	}
	if !isExist {
		if err = MinioUtil.minioClient.MakeBucket(
			ctx,
			bucketName,
			minio.MakeBucketOptions{Region: location},
		); err != nil {
			panic(err)
		}
		//设置该存储桶策略
		if err = MinioUtil.minioClient.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(read_only, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
	// 创建一个叫report的存储桶,用于放实验报告
	bucketName = "report"
	isExist, err = MinioUtil.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		panic(err)
	}
	if !isExist {
		if err = MinioUtil.minioClient.MakeBucket(
			ctx,
			bucketName,
			minio.MakeBucketOptions{Region: location},
		); err != nil {
			panic(err)
		}
		//设置该存储桶策略
		if err = MinioUtil.minioClient.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(read_only, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
}

// UploadPdf 上传pdf
// @receiver u
// @params file
// @return string
// @return error
// @date 2021-01-13 20:13:37
func (u *minioUtil) UploadPdf(file *os.File) (string, error) {
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	pdfUploadName := guuid.New().String()
	pdfUploadName = strings.ReplaceAll(pdfUploadName, "-", "")
	// 上传
	if _, err = u.minioClient.PutObject(
		context.Background(),
		reportBucketName,
		fmt.Sprintf("%s%s", pdfUploadName, ".pdf"),
		file,
		stat.Size(),
		minio.PutObjectOptions{ContentType: "application/pdf"},
	); err != nil {
		return "", err
	}
	// 返回可直接访问的url类似："http://118.178.253.239:9000/avatar/789.jpg"
	return fmt.Sprintf("http://%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		reportBucketName,
		pdfUploadName,
		".pdf",
	), nil
}

//UploadAvatar 上传头像，返回url
//@receiver u
//@params image
//@return string 头像url
//@return error
//@date 2021-01-13 18:19:48
func (u *minioUtil) UploadAvatar(file multipart.File) (string, error) {
	img, err := jpeg.Decode(file)
	if err != nil {
		return "", err
	}
	imageUploadName := guuid.New().String()
	imageUploadName = strings.ReplaceAll(imageUploadName, "-", "")
	// 压缩成缩略图
	dstImage128 := imaging.Resize(img, 128, 0, imaging.Lanczos)
	buff := new(bytes.Buffer)
	// 装入缓存
	if err = jpeg.Encode(buff, dstImage128, nil); err != nil {
		return "", err
	}
	// 上传
	if _, err = u.minioClient.PutObject(
		context.Background(),
		avatarBucketName,
		imageUploadName+"./jpg",
		buff,
		int64(buff.Len()),
		minio.PutObjectOptions{ContentType: "image/jpg"},
	); err != nil {
		return "", err
	}
	// 返回可直接访问的头像url"http://118.178.253.239:9000/avatar/789.jpg"
	return fmt.Sprintf("http://%s/%s/%s%s",
		g.Cfg().GetString("minio.endpoint"),
		avatarBucketName,
		imageUploadName,
		".jpg",
	), nil
}
