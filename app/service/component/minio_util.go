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

var (
	picBucketName        = g.Cfg().GetString("minio.bucketName.picBucketName")
	reportBucketName     = g.Cfg().GetString("minio.bucketName.reportBucketName")
	attachmentBucketName = g.Cfg().GetString("minio.bucketName.attachment")
	videoBucketName      = g.Cfg().GetString("minio.bucketName.videoBucketName")
)

const (
	// 自己试出来的，官网的sdk说明已经过时了，不能直接把"github.com/minio/minio-go/v7/pkg/policy"的常量放进去
	readOnly     = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	writeOnly    = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
	writeAndRead = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\",\"s3:GetBucketLocation\"],\"Resource\":[\"arn:aws:s3:::%s\"]},{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeletePic\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Resource\":[\"arn:aws:s3:::%s/*\"]}]}"
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
	bucketName := picBucketName
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
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(readOnly, bucketName, bucketName)); err != nil {
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
		if err = m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(readOnly, bucketName, bucketName)); err != nil {
			panic(err)
		}
	}
	return m
}
