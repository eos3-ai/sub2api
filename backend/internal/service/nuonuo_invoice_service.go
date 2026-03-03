package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const (
	nuonuoDefaultAPIURL    = "https://api.nuonuoyun.com"
	nuonuoTokenPath        = "/nuonuo/LoginByTaxNum"
	nuonuoPrintInvoicePath = "/nuonuo/printInvoice.action"
	nuonuoTokenCacheTTL    = 110 * time.Minute // refresh before 2-hour expiry
)

// NuonuoInvoiceService wraps the Nuonuo e-invoice API.
type NuonuoInvoiceService struct {
	cfg    *config.NuonuoConfig
	client *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

func NewNuonuoInvoiceService(cfg *config.NuonuoConfig) *NuonuoInvoiceService {
	if cfg == nil || !cfg.Enabled {
		return nil
	}
	apiURL := strings.TrimSpace(cfg.APIURL)
	if apiURL == "" {
		cfg.APIURL = nuonuoDefaultAPIURL
	}
	return &NuonuoInvoiceService{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// nuonuoTokenResponse is the JSON response from the Nuonuo OAuth token endpoint.
type nuonuoTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	ErrCode     string `json:"error"`
	ErrDesc     string `json:"error_description"`
}

// getAccessToken returns a cached or freshly fetched OAuth 2.0 access token.
func (s *NuonuoInvoiceService) getAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cachedToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.cachedToken, nil
	}

	apiBase := strings.TrimRight(s.cfg.APIURL, "/")
	tokenURL := apiBase + nuonuoTokenPath

	params := url.Values{}
	params.Set("client_id", s.cfg.AppKey)
	params.Set("client_secret", s.cfg.AppSecret)
	params.Set("grant_type", "client_credentials")
	params.Set("tax_no", s.cfg.TaxNo)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", fmt.Errorf("nuonuo: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("nuonuo: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("nuonuo: read token response: %w", err)
	}

	var tr nuonuoTokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("nuonuo: parse token response: %w (body: %s)", err, truncate(string(body), 200))
	}
	if tr.ErrCode != "" {
		return "", fmt.Errorf("nuonuo: token error %s: %s", tr.ErrCode, tr.ErrDesc)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("nuonuo: empty access_token in response")
	}

	ttl := nuonuoTokenCacheTTL
	if tr.ExpiresIn > 0 {
		ttl = time.Duration(tr.ExpiresIn)*time.Second - 10*time.Minute
	}
	s.cachedToken = tr.AccessToken
	s.tokenExpiry = time.Now().Add(ttl)
	return tr.AccessToken, nil
}

// NuonuoInvoiceItem represents a line item in the invoice.
type NuonuoInvoiceItem struct {
	Name     string  // 货物或劳务名称
	TaxRate  string  // 税率 e.g. "0.06"
	Quantity string  // 数量
	Unit     string  // 单位
	Price    float64 // 不含税单价
	Amount   float64 // 不含税金额
}

// NuonuoIssueRequest is the input for issuing a Nuonuo invoice.
type NuonuoIssueRequest struct {
	InvoiceType string // "normal" | "special"
	OrderNo     string // 我方申请单号，作为 serialNo

	BuyerName        string
	BuyerTaxNo       string
	BuyerAddress     string
	BuyerPhone       string
	BuyerBankName    string
	BuyerBankAccount string
	ReceiverEmail    string

	ItemName      string  // 开票内容名称
	TotalAmountCNY float64 // 含税总金额
	TaxRate       string  // 税率，默认 "0.06"
}

// NuonuoIssueResponse contains the result of a successful invoice issuance.
type NuonuoIssueResponse struct {
	SerialNo string // 诺诺流水号 (我方用于关联回调的 ID)
}

// nuonuoPrintBody is the JSON body sent to Nuonuo print invoice API.
type nuonuoPrintBody struct {
	BuyerName        string                `json:"buyerName"`
	BuyerTaxNum      string                `json:"buyerTaxNum"`
	BuyerTel         string                `json:"buyerTel"`
	BuyerAddress     string                `json:"buyerAddress"`
	BuyerBankName    string                `json:"buyerBankName"`
	BuyerBankAccount string                `json:"buyerBankAccount"`
	BuyerEmail       string                `json:"buyerEmail"`
	SalerTaxNum      string                `json:"salerTaxNum"`
	SalerTel         string                `json:"salerTel"`
	SalerAddress     string                `json:"salerAddress"`
	SalerAccount     string                `json:"salerAccount"`
	SerialNo         string                `json:"serialNo"`
	InvoiceDate      string                `json:"invoiceDate"`
	InvoiceType      string                `json:"invoiceType"` // "0"=普票, "1"=专票
	CallBackURL      string                `json:"callBackUrl"`
	InvoiceDetailList []nuonuoInvoiceItem  `json:"invoiceDetailList"`
}

type nuonuoInvoiceItem struct {
	GoodsName    string `json:"goodsName"`
	TaxRate      string `json:"taxRate"`
	TaxIncluded  string `json:"taxIncluded"` // "1"=含税
	InvoiceAmount string `json:"invoiceAmount"`
	Num          string `json:"num"`
	Unit         string `json:"unit"`
	SpecType     string `json:"specType"`
}

// nuonuoPrintResponse is the JSON response from the Nuonuo print invoice endpoint.
type nuonuoPrintResponse struct {
	Code    string                 `json:"code"`
	Describe string               `json:"describe"`
	Result  *nuonuoPrintResult     `json:"result"`
}

type nuonuoPrintResult struct {
	InvoiceSerialNum string `json:"invoiceSerialNum"`
}

// IssueInvoice calls the Nuonuo API to issue an e-invoice.
// It returns a NuonuoIssueResponse containing the Nuonuo serial number.
func (s *NuonuoInvoiceService) IssueInvoice(ctx context.Context, req *NuonuoIssueRequest) (*NuonuoIssueResponse, error) {
	if s == nil || req == nil {
		return nil, fmt.Errorf("nuonuo: service or request is nil")
	}

	token, err := s.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nuonuo: get token: %w", err)
	}

	invoiceTypeCode := "0" // 普票
	if req.InvoiceType == InvoiceTypeSpecial {
		invoiceTypeCode = "1"
	}

	taxRate := req.TaxRate
	if taxRate == "" {
		taxRate = "0.06"
	}

	// 不含税金额 = 含税金额 / (1 + 税率)
	taxRateF := 0.06
	switch taxRate {
	case "0.09":
		taxRateF = 0.09
	case "0.13":
		taxRateF = 0.13
	case "0.03":
		taxRateF = 0.03
	case "0":
		taxRateF = 0.0
	}
	var amountExcl float64
	if taxRateF > 0 {
		amountExcl = req.TotalAmountCNY / (1 + taxRateF)
	} else {
		amountExcl = req.TotalAmountCNY
	}

	body := nuonuoPrintBody{
		BuyerName:        req.BuyerName,
		BuyerTaxNum:      req.BuyerTaxNo,
		BuyerTel:         req.BuyerPhone,
		BuyerAddress:     req.BuyerAddress,
		BuyerBankName:    req.BuyerBankName,
		BuyerBankAccount: req.BuyerBankAccount,
		BuyerEmail:       req.ReceiverEmail,
		SalerTaxNum:      s.cfg.TaxNo,
		SalerAddress:     s.cfg.SellerAddress,
		SalerAccount:     s.cfg.SellerBank,
		SerialNo:         req.OrderNo,
		InvoiceDate:      time.Now().Format("2006-01-02"),
		InvoiceType:      invoiceTypeCode,
		InvoiceDetailList: []nuonuoInvoiceItem{
			{
				GoodsName:     req.ItemName,
				TaxRate:       taxRate,
				TaxIncluded:   "1",
				InvoiceAmount: fmt.Sprintf("%.2f", amountExcl),
				Num:           "1",
				Unit:          "次",
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("nuonuo: marshal request: %w", err)
	}

	apiBase := strings.TrimRight(s.cfg.APIURL, "/")
	apiURL := apiBase + nuonuoPrintInvoicePath

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("nuonuo: build print request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("accessToken", token)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nuonuo: print invoice request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("nuonuo: read print response: %w", err)
	}

	var pr nuonuoPrintResponse
	if err := json.Unmarshal(respBody, &pr); err != nil {
		return nil, fmt.Errorf("nuonuo: parse print response: %w (body: %s)", err, truncate(string(respBody), 300))
	}
	if pr.Code != "E0000" && pr.Code != "0000" {
		return nil, fmt.Errorf("nuonuo: API error code=%s describe=%s", pr.Code, pr.Describe)
	}

	serialNo := req.OrderNo
	if pr.Result != nil && pr.Result.InvoiceSerialNum != "" {
		serialNo = pr.Result.InvoiceSerialNum
	}

	log.Printf("[Nuonuo] invoice issued: serial_no=%s order_no=%s", serialNo, req.OrderNo)
	return &NuonuoIssueResponse{SerialNo: serialNo}, nil
}

// NuonuoWebhookPayload represents the callback body sent by Nuonuo.
type NuonuoWebhookPayload struct {
	// SerialNo is the internal serial number we passed when issuing.
	SerialNo   string `json:"serialNo"`
	// InvoiceCode is the national invoice code.
	InvoiceCode string `json:"fpDm"`
	// InvoiceNo is the invoice number.
	InvoiceNo  string `json:"fpHm"`
	// PDFURL is the download URL for the invoice PDF.
	PDFURL     string `json:"pdfUrl"`
	// InvoiceStatus: "2"=success, "4"=failed
	InvoiceStatus string `json:"invoiceLine"`
	// ErrMsg is set when InvoiceStatus is "4".
	ErrMsg     string `json:"errMsg"`
	// Sign is the HMAC-SHA256 signature.
	Sign       string `json:"sign"`
}

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a Nuonuo webhook.
// Nuonuo signs: HMAC-SHA256(serialNo + fpDm + fpHm, webhook_secret)
func (s *NuonuoInvoiceService) VerifyWebhookSignature(payload *NuonuoWebhookPayload) bool {
	if s == nil || payload == nil {
		return false
	}
	secret := s.cfg.WebhookSecret
	if secret == "" {
		// If no secret configured, skip verification (dev/test mode).
		return true
	}
	message := payload.SerialNo + payload.InvoiceCode + payload.InvoiceNo
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(payload.Sign))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
