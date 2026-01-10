package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// ZpayService 封装 Zpay 渠道
type ZpayService struct {
	cfg            *config.ZpayConfig
	paymentBaseURL string
}

func NewZpayService(cfg *config.Config) *ZpayService {
	var zpayCfg *config.ZpayConfig
	var paymentBaseURL string
	if cfg != nil {
		zpayCfg = &cfg.Payment.Zpay
		paymentBaseURL = cfg.Payment.BaseURL
	}
	return &ZpayService{cfg: zpayCfg, paymentBaseURL: paymentBaseURL}
}

// CreatePayment 生成 Zpay 支付链接
func (s *ZpayService) CreatePayment(ctx context.Context, order *PaymentOrder, channel string) (payURL string, qrURL string, err error) {
	if s.cfg == nil || !s.cfg.Enabled {
		return "", "", errors.New("zpay is disabled")
	}
	if order == nil {
		return "", "", errors.New("order is required")
	}
	if s.cfg.PID == "" || s.cfg.Key == "" {
		return "", "", errors.New("zpay pid/key is required")
	}

	payType := normalizeZpayPayType(channel)
	if payType == "" {
		payType = "alipay"
	}

	notifyURL, err := s.resolvePublicURL(s.cfg.NotifyURL)
	if err != nil {
		return "", "", err
	}
	returnURL, err := s.resolvePublicURL(s.cfg.ReturnURL)
	if err != nil {
		return "", "", err
	}
	if s.cfg.RequireHTTPS {
		if !strings.HasPrefix(strings.ToLower(notifyURL), "https://") || !strings.HasPrefix(strings.ToLower(returnURL), "https://") {
			return "", "", errors.New("zpay requires https notify/return url")
		}
	}

	submitURL := strings.TrimSpace(s.cfg.SubmitURL)
	if submitURL == "" {
		apiURL := strings.TrimRight(strings.TrimSpace(s.cfg.APIURL), "/")
		if apiURL == "" {
			apiURL = "https://zpayz.cn"
		}
		submitURL = apiURL + "/submit.php"
	}

	outTradeNo := order.OrderNo
	if prefix := strings.TrimSpace(s.cfg.OrderPrefix); prefix != "" {
		outTradeNo = prefix + outTradeNo
	}

	params := map[string]string{
		"pid":          strings.TrimSpace(s.cfg.PID),
		"type":         payType,
		"out_trade_no": outTradeNo,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"name":         "Recharge " + order.OrderNo,
		"money":        fmt.Sprintf("%.2f", order.AmountCNY),
	}

	// Add channel ID (cid) if configured for specific payment type
	channelID := s.selectChannelID(payType)
	if channelID != "" {
		params["cid"] = channelID
	}

	sign := zpayMD5Sign(params, s.cfg.Key)
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	values.Set("sign", sign)
	values.Set("sign_type", "MD5")

	payURL = submitURL + "?" + values.Encode()
	// For ZPay, frontend can generate QR code from pay_url directly.
	qrURL = payURL
	return payURL, qrURL, nil
}

// VerifyCallback 校验 Zpay 回调
func (s *ZpayService) VerifyCallback(ctx context.Context, data map[string]string) (orderNo string, tradeNo string, err error) {
	if s.cfg == nil || !s.cfg.Enabled {
		return "", "", errors.New("zpay is disabled")
	}
	if data == nil {
		return "", "", errors.New("callback data is required")
	}

	sign := strings.TrimSpace(data["sign"])
	if sign == "" {
		return "", "", errors.New("missing sign")
	}

	payload := make(map[string]string, len(data))
	for k, v := range data {
		if k == "sign" || k == "sign_type" {
			continue
		}
		payload[k] = v
	}

	expected := zpayMD5Sign(payload, s.cfg.Key)
	if !strings.EqualFold(expected, sign) {
		return "", "", errors.New("invalid sign")
	}

	outTradeNo := strings.TrimSpace(data["out_trade_no"])
	if outTradeNo == "" {
		outTradeNo = strings.TrimSpace(data["order_no"])
	}
	if outTradeNo == "" {
		return "", "", errors.New("missing out_trade_no")
	}

	orderNo = strings.TrimPrefix(outTradeNo, strings.TrimSpace(s.cfg.OrderPrefix))
	tradeNo = strings.TrimSpace(data["trade_no"])
	if tradeNo == "" {
		tradeNo = strings.TrimSpace(data["tradeNo"])
	}
	return orderNo, tradeNo, nil
}

func normalizeZpayPayType(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "alipay", "zpay":
		return "alipay"
	case "wechat", "wxpay":
		return "wxpay"
	default:
		return ""
	}
}

func zpayMD5Sign(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		if params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	b.WriteString(strings.TrimSpace(key))

	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func (s *ZpayService) resolvePublicURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("empty url")
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value, nil
	}
	base := strings.TrimRight(strings.TrimSpace(s.paymentBaseURL), "/")
	if base == "" {
		return "", errors.New("payment.base_url is required for relative notify/return url")
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return base + value, nil
}

// selectChannelID selects the appropriate channel ID (cid) based on payment type.
// Returns the channel ID if configured, empty string otherwise.
func (s *ZpayService) selectChannelID(payType string) string {
	if s.cfg == nil {
		return ""
	}

	switch strings.ToLower(payType) {
	case "alipay":
		return strings.TrimSpace(s.cfg.AlipayChannelID)
	case "wxpay":
		return strings.TrimSpace(s.cfg.WechatChannelID)
	default:
		return ""
	}
}
