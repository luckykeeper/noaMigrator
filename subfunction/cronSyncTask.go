// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"bytes"
	"context"
	"noaMigrator/model"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

func cronSyncTask(NoaConfig model.NoaConfig) {
	logrus.Infoln("↓------------------↓")
	logrus.Infoln("现在时间:", time.Now().Local().Format(time.RFC3339))
	logrus.Infoln("将执行一次同步任务")
	logrus.Infoln("数据库类型:", NoaConfig.DataBaseType)
	logrus.Debugln("数据库连接信息:", NoaConfig.Dsn)
	logrus.Infoln("进行同步的目录列表:", NoaConfig.SyncPath)
	logrus.Infoln("进行同步的用户 ID 列表:", NoaConfig.UserID)
	logrus.Infoln("目标 S3 存储的 Endpoint :", NoaConfig.S3Endpoint)
	logrus.Debugln("目标 S3 存储的 AccessKey :", NoaConfig.S3AccessKey)
	logrus.Debugln("目标 S3 存储的 SecretKey :", NoaConfig.S3SecretKey)
	logrus.Infoln("目标 S3 存储的存储桶:", NoaConfig.S3Bucket)
	logrus.Infoln("定时执行 Cron 表达式:", NoaConfig.Cron)

	baseDir := "./uploads/"
	var baseDirWithUID, pathToSync []string
	for _, userID := range NoaConfig.UserID {
		uIDPath := baseDir + userID
		baseDirWithUID = append(baseDirWithUID, uIDPath)
	}

	for _, path := range baseDirWithUID {
		for _, sync := range NoaConfig.SyncPath {
			syncPath := path + "/" + sync
			pathToSync = append(pathToSync, syncPath)
		}
	}
	logrus.Infoln("同步路径:", pathToSync)

	// Cloudreve 侧校验完成的文件
	fileListCloudreve := make([]model.FlieInfo, 0)

	for _, syncPath := range pathToSync {
		fileList := make([]model.FlieInfo, 0)

		err := filepath.Walk(syncPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				var thisFlieInfo model.FlieInfo
				thisFlieInfo.FileSize = uint64(info.Size())
				thisFlieInfo.FilePath = path

				if !info.IsDir() {
					thisFlieInfo.FileType, thisFlieInfo.FileMIME = CheckFileHeader(thisFlieInfo.FilePath)
					fileList = append(fileList, thisFlieInfo)
				} else {
					logrus.Infoln("跳过文件夹:", thisFlieInfo.FilePath)
				}
				return nil
			})
		if err != nil {
			logrus.Errorln(err)
		}

		logrus.Infoln("将同步文件信息如下：")
		for seq, fileInfo := range fileList {
			logrus.Infoln(seq+1, "\t",
				fileInfo.FilePath, "\t",
				fileInfo.FileSize, "\t",
				fileInfo.FileType, "\t",
				fileInfo.FileMIME)
		}

		logrus.Infoln("开始校验数据库信息!")
		getPolicy := model.Policies{Type: "local"}
		CocoaDataEngine.Where("type = ?", getPolicy.Type).Get(&getPolicy)
		// 注意存储策略需要是本地，存储策略是本地的必须只有一个
		logrus.Infoln("本地存储策略的策略 ID:", getPolicy)

		for _, fileInfo := range fileList {
			var getCloudreveInfo model.FlieInfo
			var path string
			if runtime.GOOS == "windows" {
				path = strings.ReplaceAll(fileInfo.FilePath, "\\", "/")
			} else {
				path = fileInfo.FilePath
			}
			CocoaDataEngine.Table("files").Cols("size", "policy_id").Where("source_name=?", path).Get(&getCloudreveInfo)
			getCloudreveInfo.FilePath = fileInfo.FilePath
			logrus.Infoln("数据库内找到的匹配信息:", getCloudreveInfo.FilePath, "\t",
				getCloudreveInfo.Size, "\t",
				getCloudreveInfo.PolicyId)

			if (fileInfo.FileSize == getCloudreveInfo.Size) && (getCloudreveInfo.PolicyId == getPolicy.Id) {
				logrus.Infoln("存储策略和文件大小校验通过:", fileInfo.FilePath)
				fileListCloudreve = append(fileListCloudreve, fileInfo)
			} else {
				logrus.Warnln("寄！寄！寄啦！存储策略和文件大小校验未通过:", fileInfo.FilePath)
				logrus.Warnln("fileInfo.Size:", fileInfo.Size)
				logrus.Warnln("getCloudreveInfo.Size:", getCloudreveInfo.Size)
				logrus.Warnln("getCloudreveInfo.PolicyId:", getCloudreveInfo.PolicyId)
				logrus.Warnln("getPolicy.Id:", getPolicy.Id)
				logrus.Errorln("杂❤鱼❤酱❤~快来看看文件出什么问题啦,zako~❤zako~❤")
			}
		}
	}

	var fileListCloudreveWithMD5 []model.FlieInfo

	// 校验通过的文件，计算 MD5
	for _, fileInfo := range fileListCloudreve {
		var err error
		fileInfo.CheckSumMD5, err = GetMD5FromFile(fileInfo.FilePath)
		if err != nil {
			logrus.Errorln("计算文件MD5失败:", fileInfo.FilePath, err)
			continue
		}
		logrus.Debugln("fileInfo.FilePath:", fileInfo.FilePath, "MD5:", fileInfo.CheckSumMD5)
		fileListCloudreveWithMD5 = append(fileListCloudreveWithMD5, fileInfo)
	}

	// S3 侧的文件
	fileListS3 := make([]model.FlieInfo, 0)

	// 获取协议
	protocol := strings.Split(NoaConfig.S3Endpoint, "://")[0]
	endpoint := strings.Split(NoaConfig.S3Endpoint, "://")[1]

	var useSSL bool
	if protocol == "https" {
		useSSL = true
	} else if protocol == "http" {
		useSSL = false
	} else {
		logrus.Errorln("S3 配置有误")
		return
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(NoaConfig.S3AccessKey, NoaConfig.S3SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Errorln("创建 MinIO 客户端失败", err)
		return
	}
	logrus.Infoln("创建 MinIO 客户端成功")

	opts := minio.ListObjectsOptions{
		// 去除路径开头的 /
		Recursive: true,
	}

	// 建立临时校验文件夹

	err = os.Mkdir("./temp", 0644)
	if err != nil {
		logrus.Infoln("临时文件夹已经存在，跳过创建")
	}

	for object := range minioClient.ListObjects(context.Background(), NoaConfig.S3Bucket, opts) {
		if object.Err != nil {
			logrus.Errorln(object.Err)
			return
		}
		thisS3File := model.FlieInfo{FilePath: object.Key, CheckSumMD5: object.ETag}

		logrus.Infoln("下载文件校验MD5:", NoaConfig.S3Endpoint+"/"+NoaConfig.S3Bucket+"/"+thisS3File.FilePath)
		cocoaDownResult := cocoaTryDownload("./temp/tempFile", NoaConfig.S3Endpoint+"/"+NoaConfig.S3Bucket+"/"+thisS3File.FilePath)
		if cocoaDownResult {
			logrus.Infoln("文件下载成功")
			thisS3File.CheckSumMD5, err = GetMD5FromFile("./temp/tempFile")
			if err != nil {
				logrus.Errorln("计算MD5失败:", err)
			} else {
				logrus.Infoln("文件MD5:", thisS3File.CheckSumMD5)
			}
		} else {
			logrus.Errorln("文件下载失败")
		}
		err = os.Remove("./temp/tempFile")
		if err != nil {
			logrus.Errorln("移除临时文件失败:", err)
		}

		fileListS3 = append(fileListS3, thisS3File)
	}
	logrus.Debugln("!!!")
	logrus.Debugln("fileListCloudreveWithMD5: ", fileListCloudreveWithMD5)
	logrus.Debugln("fileListS3: ", fileListS3)

	var cloudreveUploadList, s3RemoveFileList []model.FlieInfo

	for _, fileInfo := range fileListCloudreveWithMD5 {
		var comparePath string
		matchTag := false
		if runtime.GOOS == "windows" {
			comparePath = strings.SplitN(fileInfo.FilePath, "\\", 3)[2]
			comparePath = strings.ReplaceAll(comparePath, "\\", "/")
		} else {
			comparePath = strings.SplitN(fileInfo.FilePath, "/", 3)[2]
		}
		for _, fileS3 := range fileListS3 {
			if comparePath == fileS3.FilePath {
				if fileInfo.CheckSumMD5 == fileS3.CheckSumMD5 {
					logrus.Infoln("匹配成功:", fileInfo.FilePath)
					matchTag = true
					break
					// fileListCloudreveWithMD5 = removeItemFromFileInfo(fileListCloudreveWithMD5, fileInfo.FilePath)
					// fileListS3 = removeItemFromFileInfo(fileListS3, fileS3.FilePath)
					// }
				} else {
					logrus.Infoln("匹配失败:", fileInfo.FilePath)
					logrus.Infoln("fileInfo comparePath:", comparePath)
					logrus.Infoln("fileInfo MD5:", fileInfo.CheckSumMD5)
					logrus.Infoln("S3 FilePath:", fileS3.FilePath)
					logrus.Infoln("S3 MD5:", fileS3.CheckSumMD5)
				}
			}

		}
		if !matchTag {
			cloudreveUploadList = append(cloudreveUploadList, fileInfo)
		}
	}

	for _, fileS3 := range fileListS3 {
		matchTag := false
		for _, fileInfo := range fileListCloudreveWithMD5 {
			var comparePath string
			if runtime.GOOS == "windows" {
				comparePath = strings.SplitN(fileInfo.FilePath, "\\", 3)[2]
				comparePath = strings.ReplaceAll(comparePath, "\\", "/")
			} else {
				comparePath = strings.SplitN(fileInfo.FilePath, "/", 3)[2]
			}
			if comparePath == fileS3.FilePath {
				if fileInfo.CheckSumMD5 == fileS3.CheckSumMD5 {
					logrus.Infoln("匹配成功:", fileInfo.FilePath)
					matchTag = true
					break
					// fileListCloudreveWithMD5 = removeItemFromFileInfo(fileListCloudreveWithMD5, fileInfo.FilePath)
					// fileListS3 = removeItemFromFileInfo(fileListS3, fileS3.FilePath)
					// }
				} else {
					logrus.Infoln("匹配失败:", fileInfo.FilePath)
					logrus.Infoln("fileInfo comparePath:", comparePath)
					logrus.Infoln("fileInfo MD5:", fileInfo.CheckSumMD5)
					logrus.Infoln("S3 FilePath:", fileS3.FilePath)
					logrus.Infoln("S3 MD5:", fileS3.CheckSumMD5)
				}
			}
		}
		if !matchTag {
			s3RemoveFileList = append(s3RemoveFileList, fileS3)
		}
	}
	logrus.Debugln("-----------")
	logrus.Debugln("将移除的文件列表(S3):", s3RemoveFileList)
	logrus.Debugln("将同步的文件列表(cloudreve):", cloudreveUploadList)
	logrus.Debugln("-----------")

	logrus.Infoln("同步开始!")
	if len(fileListS3) > 0 {
		logrus.Warnln("开始移除 S3 侧未匹配的文件:", s3RemoveFileList)
		for _, fileInfo := range s3RemoveFileList {
			S3Remover(fileInfo.FilePath, NoaConfig)
		}
		logrus.Infoln("移除所有 S3 侧未匹配的文件完成")
	}

	if len(cloudreveUploadList) > 0 {
		logrus.Infoln("同步所有 cloudreve 侧匹配的文件")
		for _, fileInfo := range cloudreveUploadList {
			var objectName string
			if runtime.GOOS == "windows" {
				objectName = strings.SplitN(fileInfo.FilePath, "\\", 3)[2]
				objectName = strings.ReplaceAll(objectName, "\\", "/")
			} else {
				objectName = strings.SplitN(fileInfo.FilePath, "/", 3)[2]
			}
			file, err := os.ReadFile(fileInfo.FilePath)
			if err != nil {
				logrus.Errorln("os.ReadFile:", fileInfo.FilePath, err)
			}
			reader := bytes.NewReader(file)
			s3Url, result := S3Uploader(objectName, fileInfo.FileMIME, reader, int64(fileInfo.FileSize), NoaConfig)
			if !result {
				logrus.Panicln("上传到 S3 存储失败")
			}
			logrus.Infoln("上传成功,路径:", s3Url)
		}
	}
	logrus.Infoln("生成Excel->Start")
	generateExcel(NoaConfig)
	logrus.Infoln("生成Excel->End")
	logrus.Infoln("本周期同步完成!")

	logrus.Infoln("↑------------------↑")
}

// func removeItemFromFileInfo(fileInfoList []model.FlieInfo, removeFilePath string) []model.FlieInfo {
// 	var fileInfoListReturn []model.FlieInfo
// 	// logrus.Debugln("移除前的清单: ", fileInfoListReturn)
// 	for _, fileInfo := range fileInfoList {
// 		if removeFilePath != fileInfo.FilePath {
// 			fileInfoListReturn = append(fileInfoListReturn, fileInfo)
// 		} // } else {
// 		// logrus.Debugln("移除匹配项: ", fileInfo.FilePath)
// 		// }
// 	}
// 	// logrus.Debugln("移除后的清单: ", fileInfoListReturn)
// 	return fileInfoListReturn
// }
