package service

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

const baiduOcpcURL = "https://ocpc.baidu.com/ocpcapi/api/uploadConvertData"

// ReportBaiduOcpcEvent 异步上报百度 OCPC 转化事件（goroutine 执行，不阻塞业务）
// newType: 1=关键页面浏览 3=注册 5=登录
func ReportBaiduOcpcEvent(bdVid, landingUrl string, newType int) {
	token := os.Getenv("BAIDU_OCPC_TOKEN")
	if token == "" || bdVid == "" {
		return
	}
	logidUrl := landingUrl
	if logidUrl == "" {
		logidUrl = "https://placeholder.invalid?bd_vid=" + bdVid
	}
	go func() {
		body, err := json.Marshal(map[string]any{
			"token": token,
			"conversionTypes": []map[string]any{
				{"logidUrl": logidUrl, "newType": newType},
			},
		})
		if err != nil {
			slog.Warn("baidu ocpc marshal error", "err", err)
			return
		}
		resp, err := http.Post(baiduOcpcURL, "application/json", bytes.NewReader(body))
		if err != nil {
			slog.Warn("baidu ocpc request error", "err", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Warn("baidu ocpc non-200 response", "status", resp.StatusCode)
		}
	}()
}
