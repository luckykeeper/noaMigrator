// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"noaMigrator/model"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

// CronTaskCenter 定时器
var CronTaskCenter *cron.Cron

// 定时任务入口
func CronTaskInit(NoaConfig model.NoaConfig) {
	logrus.Println("Initializing CronCenter During Early Start!")
	CronTaskCenter = cron.New() // 定时任务
	CronTaskCenter.Start()

	CronTaskCenter.AddFunc("@every 60s", cronCheckDatabase)

	CronTaskCenter.AddFunc(NoaConfig.Cron, func() { cronSyncTask(NoaConfig) })

}
