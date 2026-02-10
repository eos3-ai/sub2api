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

// NewDingtalkService creates a DingTalk notification client.
//
// DingTalk settings are env/config-file driven (see config bindings), not stored in DB.
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
	if s == nil {
		return false
	}
	return s.cfg != nil && s.cfg.Enabled && strings.TrimSpace(s.cfg.WebhookURL) != ""
}

// EnabledForPayment indicates whether payment/recharge order notifications should be sent.
//
// It is intentionally separate from Enabled(), so operational alerts can be enabled
// while payment notifications are disabled.
func (s *DingtalkService) EnabledForPayment() bool {
	if s == nil {
		return false
	}
	if !s.Enabled() {
		return false
	}
	return s.cfg != nil && s.cfg.PaymentNotifyEnabled
}

func (s *DingtalkService) SendMarkdown(ctx context.Context, title string, text string) error {
	if s == nil {
		return nil
	}
	return s.SendMarkdownWithConfig(ctx, s.cfg, title, text)
}

// SendMarkdownWithConfig sends a markdown message using the provided config.
// It does NOT read settings from DB and is intended for admin "test alert" calls.
func (s *DingtalkService) SendMarkdownWithConfig(ctx context.Context, cfg *config.DingtalkConfig, title string, text string) error {
	if s == nil || cfg == nil || !cfg.Enabled || strings.TrimSpace(cfg.WebhookURL) == "" {
		return nil
	}

	title, text = applyDingtalkEnvTag(cfg, title, text)

	endpoint, err := signedDingtalkWebhookURL(strings.TrimSpace(cfg.WebhookURL), strings.TrimSpace(cfg.Secret))
	if err != nil {
		return err
	}

	atMobiles := parseCommaSeparated(cfg.AtMobiles)

	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"title": title,
			"text":  text,
		},
		"at": map[string]any{
			"atMobiles": atMobiles,
			"isAtAll":   cfg.AtAll,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := s.client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func applyDingtalkEnvTag(cfg *config.DingtalkConfig, title string, text string) (string, string) {
	if cfg == nil {
		return title, text
	}

	env := strings.TrimSpace(cfg.Env)
	if env == "" {
		return title, text
	}
	env = sanitizeDingtalkInlineCode(env)
	if env == "" {
		return title, text
	}

	trimmedTitle := strings.TrimSpace(title)
	prefix := fmt.Sprintf("【%s】", env)
	if trimmedTitle == "" {
		title = prefix
	} else if !strings.HasPrefix(trimmedTitle, prefix) {
		title = prefix + trimmedTitle
	}

	envLine := fmt.Sprintf("**环境**：`%s`  \n\n", env)
	if strings.TrimSpace(text) == "" {
		text = envLine
	} else if !strings.HasPrefix(text, envLine) {
		text = envLine + text
	}

	return title, text
}

func sanitizeDingtalkInlineCode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"`", "'",
	)
	return replacer.Replace(s)
}

func signedDingtalkWebhookURL(raw string, secret string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("missing dingtalk webhook url")
	}
	secret = strings.TrimSpace(secret)
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
	return base64.StdEncoding.EncodeToString(sum)
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
