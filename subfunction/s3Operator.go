// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"context"
	"fmt"
	"io"
	"noaMigrator/model"
	"os"
	"strings"

	"github.com/h2non/filetype"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

func S3BucketChecker(NoaConfig model.NoaConfig) (result bool) {
	ctx := context.Background()
	// 获取协议
	protocol := strings.Split(NoaConfig.S3Endpoint, "://")[0]
	endpoint := strings.Split(NoaConfig.S3Endpoint, "://")[1]

	var useSSL bool
	if protocol == "https" {
		useSSL = true
	} else if protocol == "http" {
		useSSL = false
	} else {
		return false
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(NoaConfig.S3AccessKey, NoaConfig.S3SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Errorln("创建 MinIO 客户端失败", err)
		return false
	}
	logrus.Infoln("创建 MinIO 客户端成功")

	exists, errBucketExists := minioClient.BucketExists(ctx, NoaConfig.S3Bucket)
	if errBucketExists == nil && exists {
		logrus.Infof("进入存储桶 %s 成功", NoaConfig.S3Bucket)
		return true
	} else {
		logrus.Fatalln("存储桶不存在,请先创建存储桶:", err)
		return false
	}
}

func S3Uploader(objectName, fileMIME string, fileReader io.Reader, fileSize int64, NoaConfig model.NoaConfig) (s3Url string, result bool) {
	ctx := context.Background()
	// 获取协议
	protocol := strings.Split(NoaConfig.S3Endpoint, "://")[0]
	endpoint := strings.Split(NoaConfig.S3Endpoint, "://")[1]

	var useSSL bool
	if protocol == "https" {
		useSSL = true
	} else if protocol == "http" {
		useSSL = false
	} else {
		return "", false
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(NoaConfig.S3AccessKey, NoaConfig.S3SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Errorln("创建 MinIO 客户端失败", err)
		return "", false
	}
	logrus.Infoln("创建 MinIO 客户端成功")

	exists, errBucketExists := minioClient.BucketExists(ctx, NoaConfig.S3Bucket)
	if errBucketExists == nil && exists {
		logrus.Infof("进入存储桶 %s 成功", NoaConfig.S3Bucket)
	} else {
		logrus.Errorln("存储桶不存在", err)
		return "", false
	}

	logrus.Infoln("将上传:", objectName)

	// 上传方法是 PUT ，需要在 WAF 上面关闭 PUT 方法的缓存
	// 建议关闭所有方法的缓存，否则文件可能无法上传（WAF 缓存了结果，上传文件不仅仅用了 PUT 方法）
	// log.Println("____")
	// log.Println("bucket:", bucket)
	// log.Println("objectName:", objectName)
	// log.Println("filePath:", filePath)
	// log.Println("____")
	// info, err := minioClient.PutObject(ctx, bucket, objectName, fileReader, -1, minio.PutObjectOptions{ContentType: fileMIME})
	info, err := minioClient.PutObject(ctx, NoaConfig.S3Bucket, objectName, fileReader, fileSize, minio.PutObjectOptions{ContentType: fileMIME})
	if err != nil {
		logrus.Errorln("上传失败：", err)
		return "", false
	}

	logrus.Infoln("成功上传文件： " + objectName + " ，其大小为： " + fmt.Sprint(info.Size))

	s3ObjectUrl := NoaConfig.S3Endpoint + "/" + NoaConfig.S3Bucket + "/" + objectName
	logrus.Infoln("文件链接：", s3ObjectUrl)
	return s3ObjectUrl, true
}

func S3Remover(objectName string, NoaConfig model.NoaConfig) (result bool) {
	ctx := context.Background()
	// 获取协议
	protocol := strings.Split(NoaConfig.S3Endpoint, "://")[0]
	endpoint := strings.Split(NoaConfig.S3Endpoint, "://")[1]

	var useSSL bool
	if protocol == "https" {
		useSSL = true
	} else if protocol == "http" {
		useSSL = false
	} else {
		return false
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(NoaConfig.S3AccessKey, NoaConfig.S3SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Errorln("创建 MinIO 客户端失败", err)
		return false
	}
	logrus.Infoln("创建 MinIO 客户端成功")

	exists, errBucketExists := minioClient.BucketExists(ctx, NoaConfig.S3Bucket)
	if errBucketExists == nil && exists {
		logrus.Infof("进入存储桶 %s 成功", NoaConfig.S3Bucket)
	} else {
		logrus.Errorln("存储桶不存在", err)
		return false
	}

	err = minioClient.RemoveObject(ctx, NoaConfig.S3Bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		logrus.Errorln("删除失败：", err)
		return false
	}

	logrus.Infoln("成功删除文件： " + objectName)

	return true
}

// 检查文件头
func CheckFileHeader(filePath string) (fileType, fileMIME string) {
	buf, _ := os.ReadFile(filePath)
	kind, _ := filetype.Match(buf)
	fileType = kind.Extension
	fileMIME = kind.MIME.Value

	return
}
