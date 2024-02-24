// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/sirupsen/logrus"
)

func cocoaTryDownload(destination, url string) bool {

	// ignore https cert check
	// https://github.com/cavaliergopher/grab/issues/17
	// https://github.com/cavaliergopher/grab/issues/79

	client := grab.NewClient()
	client.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	req, err := grab.NewRequest(destination, url)
	req.HTTPRequest.Header.Del("User-Agent")
	req.HTTPRequest.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edge/120.0.0.0")

	logrus.Debugln("[cocoaTryDownload]req.HTTPRequest.Header-User-Agent:", req.HTTPRequest.Header.Get("User-Agent"))

	if err != nil {
		logrus.Errorln("[cocoaTryDownload]下载失败:", err)
		return false
	}

	resp := client.Do(req)

	// start UI loop
	t := time.NewTicker(1000 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			var progress string
			progress = fmt.Sprint(100 * resp.Progress())
			if len(progress) >= 6 {
				progress = progress[0:5]
			}
			text := "  下载进度: " + fmt.Sprint(resp.BytesComplete()/1024/1024) + " / " + fmt.Sprint(resp.Size()/1024/1024) + "MB | (完成:" + progress + "%)"
			logrus.Infoln(text)

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		return false
	}
	return true
}
