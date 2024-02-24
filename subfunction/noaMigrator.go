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
	"net/http"
	"noaMigrator/model"
	ushioNoaRouter "noaMigrator/router"

	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	AppVersion = "v1.0.0_Build20240209_By_LuckyKeeper<https://github.com/luckykeeper | https://luckykeeper.site | luckykeeper@luckykeeper.site>"
)

func StartNoaMigrator(NoaConfig model.NoaConfig, debugMode bool) {
	logrus.Infoln("启动准备完成，开始校验配置项...")
	logrus.Infoln("开始校验 S3 配置")
	bucketCheckResult := S3BucketChecker(NoaConfig)
	if !bucketCheckResult {
		logrus.Fatalln("存储桶配置可能有误，检查 S3 配置及 S3 服务是否正常运行")
	}
	logrus.Infoln("S3 配置校验成功")
	logrus.Infoln("程序开始运行...")
	startAPIServer(NoaConfig, debugMode)
}

func startAPIServer(NoaConfig model.NoaConfig, devMode bool) {
	if devMode {
		logrus.Warnln("以调试模式启动 noaMigrator !")
	} else {
		logrus.Infoln("以生产环境启动 noaMigrator !")
	}
	if !devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	ushioNoaRouter.UshioNoaRouter = gin.Default()
	if devMode {
		ushioNoaRouter.GinRouter(ushioNoaRouter.UshioNoaRouter, true, AppVersion)
	} else {
		ushioNoaRouter.GinRouter(ushioNoaRouter.UshioNoaRouter, false, AppVersion)
	}

	gin.ForceConsoleColor()

	// 使用反代
	ushioNoaRouter.UshioNoaRouter.ForwardedByClientIP = true

	srv := &http.Server{
		Addr:    ":" + NoaConfig.APIPort,
		Handler: ushioNoaRouter.UshioNoaRouter,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalln("listen: ", err)
		}

	}()

	// Gracefully Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Infoln("接收到中止信号,等待最长15秒处理完剩余连接后关闭服务")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalln("服务器关闭超时，强制退出，原因:", err)
	}
	logrus.Infoln("noaMigrator Gracefully Shutdown!")

}
