// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/24 16:54
// @LastModified	: 2024/02/24 16:54
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"noaMigrator/model"
	"os"
	"strconv"
	"strings"
)

func generateExcel(config model.NoaConfig) {
	cocoaExcel := excelize.NewFile()
	defer func() {
		if err := cocoaExcel.Close(); err != nil {
			logrus.Errorln(err)
		}
	}()
	for _, user := range config.UserID {
		// 创建一个工作表
		index, err := cocoaExcel.NewSheet(user)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		// 设置单元格的值
		err = cocoaExcel.DeleteSheet("Sheet1")
		if err != nil {
			logrus.Warnln("移除默认 sheet 失败:", err)
		}
		// 设置工作簿的默认工作表
		cocoaExcel.SetActiveSheet(index)

		// 拉取 Cloudreve 数据
		cloudReveData := make([]model.CloudreveFile, 0)
		err = CocoaDataEngine.Table("files").Where("user_id = ?", user).Find(&cloudReveData)
		if err != nil {
			logrus.Errorln("拉取 Cloudreve 数据失败:", err)
			return
		}
		cocoaExcel.SetCellValue(user, "A1", "文件ID")
		cocoaExcel.SetCellValue(user, "B1", "文件名")
		cocoaExcel.SetCellValue(user, "C1", "原文件存储位置")
		cocoaExcel.SetCellValue(user, "D1", "S3存储位置")
		cocoaExcel.SetCellValue(user, "E1", "用户ID")

		cutPathSring := "uploads/" + user
		for seq, data := range cloudReveData {
			fileAbsPath := strings.Split(data.SourceName, cutPathSring)[1]
			data.S3Path = config.S3Endpoint + "/" + config.S3Bucket + fileAbsPath

			cocoaExcel.SetCellValue(user, "A"+strconv.Itoa(seq+2), data.Id)
			cocoaExcel.SetCellValue(user, "B"+strconv.Itoa(seq+2), data.Name)
			cocoaExcel.SetCellValue(user, "C"+strconv.Itoa(seq+2), data.SourceName)
			cocoaExcel.SetCellValue(user, "D"+strconv.Itoa(seq+2), data.S3Path)
			cocoaExcel.SetCellValue(user, "E"+strconv.Itoa(seq+2), data.UserId)
		}

	}

	err := os.Remove("./noaMigrator.xlsx")
	if err != nil {
		logrus.Infoln("没有发现旧的导出文件")
	}

	// 根据指定路径保存文件
	if err := cocoaExcel.SaveAs("./noaMigrator.xlsx"); err != nil {
		logrus.Errorln(err)
		return
	}

}
