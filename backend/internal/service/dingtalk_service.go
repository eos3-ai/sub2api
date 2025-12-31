package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type DingtalkService struct {
	cfg    *config.DingtalkConfig
	client *http.Client
}

func NewDingtalkService(cfg *config.Config) *DingtalkService {
	var dc *config.DingtalkConfig
	if cfg != nil {
		dc = &cfg.Dingtalk
	}
	return &DingtalkService{
		cfg: dc,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *DingtalkService) Enabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Enabled && strings.TrimSpace(s.cfg.WebhookURL) != ""
}

func (s *DingtalkService) SendMarkdown(ctx context.Context, title string, text string) error {
	if !s.Enabled() {
		return nil
	}
	endpoint, err := s.signedWebhookURL()
	if err != nil {
		return err
	}

	atMobiles := parseCommaSeparated(s.cfg.AtMobiles)

	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"title": title,
			"text":  text,
		},
		"at": map[string]any{
			"atMobiles": atMobiles,
			"isAtAll":   s.cfg.AtAll,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *DingtalkService) signedWebhookURL() (string, error) {
	raw := strings.TrimSpace(s.cfg.WebhookURL)
	if raw == "" {
		return "", fmt.Errorf("missing dingtalk webhook url")
	}
	secret := strings.TrimSpace(s.cfg.Secret)
	if secret == "" {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	ts := time.Now().UnixMilli()
	signature := dingtalkSign(ts, secret)
	q := u.Query()
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", signature)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func dingtalkSign(timestampMillis int64, secret string) string {
	message := fmt.Sprintf("%d\n%s", timestampMillis, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	sum := mac.Sum(nil)
	return url.QueryEscape(base64.StdEncoding.EncodeToString(sum))
}

func parseCommaSeparated(v string) []string {
	raw := strings.TrimSpace(v)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

