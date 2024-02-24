// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package router

import (
	"net/http"
	"noaMigrator/handler"

	"github.com/gin-gonic/gin"
)

var UshioNoaRouter *gin.Engine

func GinRouter(router *gin.Engine, devMode bool, Version string) {
	// ------------------------------------------
	// noaMigrator - 根目录 - 路由组
	serviceRootRouter := router.Group("/")

	// ------------------------------------------
	// noaMigrator - 根目录 - 跳转到 V1
	serviceRootRouter.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/v1/")
	})

	// ------------------------------------------
	// noaMigrator - 存活与健康检测 - 路由组
	checkPointRouter := serviceRootRouter.Group("/checkpoint")

	// 存活检测
	checkPointRouter.GET("/ready", func(ctx *gin.Context) {
		handler.ServiceReady(ctx, Version)
	})
	// 健康检测
	checkPointRouter.GET("/liveness", func(ctx *gin.Context) {
		handler.ServiceReady(ctx, Version)
	})

	// ------------------------------------------
	// noaMigrator - V1 - 路由组
	// v1 := serviceRootRouter.Group("/v1")

	// v1.POST("/getDeploymentApps", handler.GetDeploymentApps)
}
