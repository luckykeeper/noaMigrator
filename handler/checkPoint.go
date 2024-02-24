// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package handler

import (
	"noaMigrator/model"

	"github.com/gin-gonic/gin"
)

// 数据库测试结果，定时校验，成功200，失败500
var DataBaseCheckResult = 200

// 存活检测
func ServiceReady(context *gin.Context, Version string) {
	// logrus.Debugln("存活探针检测开始")
	var serviceStatus model.CheckPoint
	serviceStatus.AppName = "noaMigrator"
	serviceStatus.StatusCode = DataBaseCheckResult
	serviceStatus.Version = Version
	context.JSON(DataBaseCheckResult, serviceStatus)
	// logrus.Debugln("存活探针检测正常结束")
}
