package service

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const baiduOcpcURL = "https://ocpc.baidu.com/ocpcapi/api/uploadConvertData"

const (
	DefaultBaiduOcpcLandingNewType  = 1
	DefaultBaiduOcpcRegisterNewType = 3
	DefaultBaiduOcpcLoginNewType    = 5
)

func getBaiduOcpcNewType(envKey string, defaultValue int) int {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("invalid baidu ocpc new_type config", "env", envKey, "value", raw, "default", defaultValue)
		return defaultValue
	}

	return value
}

func GetBaiduOcpcLandingNewType() int {
	return getBaiduOcpcNewType("BAIDU_OCPC_NEW_TYPE_LANDING", DefaultBaiduOcpcLandingNewType)
}

func GetBaiduOcpcRegisterNewType() int {
	return getBaiduOcpcNewType("BAIDU_OCPC_NEW_TYPE_REGISTER", DefaultBaiduOcpcRegisterNewType)
}

func GetBaiduOcpcLoginNewType() int {
	return getBaiduOcpcNewType("BAIDU_OCPC_NEW_TYPE_LOGIN", DefaultBaiduOcpcLoginNewType)
}

// ReportBaiduOcpcEvent 异步上报百度 OCPC 转化事件（goroutine 执行，不阻塞业务）
// newType: 以百度后台配置为准（默认：1=关键页面浏览 3=注册 5=登录）
func ReportBaiduOcpcEvent(bdVid, landingUrl string, newType int) {
	token := os.Getenv("BAIDU_OCPC_TOKEN")
	if token == "" || bdVid == "" {
		slog.Info("baidu ocpc skip", "reason", map[string]bool{"no_token": token == "", "no_vid": bdVid == ""}, "new_type", newType)
		return
	}
	slog.Info("baidu ocpc reporting", "new_type", newType, "bd_vid", bdVid[:min(8, len(bdVid))]+"...")
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
		} else {
			slog.Info("baidu ocpc reported ok", "new_type", newType)
		}
	}()
}
