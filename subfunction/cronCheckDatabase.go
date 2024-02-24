// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"noaMigrator/handler"

	"github.com/sirupsen/logrus"
)

func cronCheckDatabase() {
	err := CocoaDataEngine.Ping()
	if err != nil {
		logrus.Errorln("数据库连接健康检测失败！", err)
		handler.DataBaseCheckResult = 500
	} else {
		handler.DataBaseCheckResult = 200
	}
}
