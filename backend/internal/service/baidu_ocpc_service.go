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

// ReportBaiduOcpcEvent reports a Baidu OCPC conversion asynchronously. Missing
// token or bd_vid is treated as a no-op so auth flows never depend on tracking.
func ReportBaiduOcpcEvent(bdVid, landingURL string, newType int) {
	token := strings.TrimSpace(os.Getenv("BAIDU_OCPC_TOKEN"))
	bdVid = strings.TrimSpace(bdVid)
	if token == "" || bdVid == "" {
		slog.Info("baidu ocpc skip", "no_token", token == "", "no_vid", bdVid == "", "new_type", newType)
		return
	}

	logIDURL := strings.TrimSpace(landingURL)
	if logIDURL == "" {
		logIDURL = "https://placeholder.invalid?bd_vid=" + bdVid
	}

	go func() {
		body, err := json.Marshal(map[string]any{
			"token": token,
			"conversionTypes": []map[string]any{
				{"logidUrl": logIDURL, "newType": newType},
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
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			slog.Warn("baidu ocpc non-200 response", "status", resp.StatusCode)
			return
		}
		slog.Info("baidu ocpc reported ok", "new_type", newType)
	}()
}
