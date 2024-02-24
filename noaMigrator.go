// noaMigrator - cloudreve -> minio 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package main

import (
	"noaMigrator/model"
	"noaMigrator/subfunction"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

// 基础变量设置
var (
	NoaConfig model.NoaConfig
)

func noaMigratorCLI() {
	noaMigrator := &cli.App{
		Name: "noaMigrator",
		Usage: "noaMigrator - cloudreve -> minio 同步推送工具" +
			"\nPowered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>" +
			"\n————————————————————————————————————————" +
			"\n注意：使用前需要先填写同目录下 config.yaml !" +
			"\n依赖：Chrome",
		Version: "1.0.0_build20240207",
		Commands: []*cli.Command{
			{
				Name:    "runProduction",
				Aliases: []string{"r"},
				Usage:   "启动同步迁移工具",
				Action: func(cCtx *cli.Context) error {
					noaMigrator(false)
					return nil
				},
			},

			{
				Name:    "runDebug",
				Aliases: []string{"rd"},
				Usage:   "启动同步迁移工具（调试模式）",
				Action: func(cCtx *cli.Context) error {
					noaMigrator(true)
					return nil
				},
			},
		},
		Copyright: "Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site> | https://github.com/luckykeeper",
	}

	if err := noaMigrator.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

// 程序入口
func main() {
	noaMigratorCLI()
}

func noaMigrator(debugMode bool) {

	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     true,
	})
	// Debug:5 | Info:4 | Warn: 3 | Error:2 | Fatal:1
	// 选定某个日志级别时，该级别及小于这个数字的级别的日志将会显示
	if debugMode {
		logrus.SetLevel(logrus.Level(5))
		logrus.Debugln("Run As Debug Mode!!!")
	} else {
		logrus.SetLevel(logrus.Level(4))
	}

	readConfig()

	logrus.Infoln("连接到 Cloudreve 数据库...")
	subfunction.InitializeDatabase(NoaConfig, debugMode)
	logrus.Infoln("连接到 Cloudreve 数据库并校验完成")
	logrus.Infoln("初始化定时任务")
	subfunction.CronTaskInit(NoaConfig)
	logrus.Infoln("定时任务初始化完成")
	subfunction.StartNoaMigrator(NoaConfig, debugMode)
}

// read config.yaml
func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	// 校验配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Fatalln("没有找到配置文件，请检查 ./config.yaml 是否存在！")
		} else {
			logrus.Fatalln("配置文件校验失败，请检查 ./config.yaml 是否存在语法错误")
		}
	}
	// 如果需要嵌套解析 yaml ，需要把 struct 也做成嵌套的
	// 参考：https://blog.csdn.net/weixin_42586723/article/details/121162029
	viper.Unmarshal(&NoaConfig)

	logrus.Infoln("读取配置成功，当前配置如下：")
	logrus.Println("____________________________")
	logrus.Infoln("数据库类型:", NoaConfig.DataBaseType)
	logrus.Debugln("数据库连接信息:", NoaConfig.Dsn)
	logrus.Infoln("进行同步的目录列表:", NoaConfig.SyncPath)
	logrus.Infoln("进行同步的用户 ID 列表:", NoaConfig.UserID)
	logrus.Infoln("目标 S3 存储的 Endpoint :", NoaConfig.S3Endpoint)
	logrus.Debugln("目标 S3 存储的 AccessKey :", NoaConfig.S3AccessKey)
	logrus.Debugln("目标 S3 存储的 SecretKey :", NoaConfig.S3SecretKey)
	logrus.Infoln("目标 S3 存储的存储桶:", NoaConfig.S3Bucket)
	logrus.Infoln("定时执行 Cron 表达式:", NoaConfig.Cron)
	logrus.Println("____________________________")
}
