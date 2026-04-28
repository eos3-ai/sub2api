package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const dingtalkMaxTimeDiff = time.Hour

var (
	dingtalkAtUserRe   = regexp.MustCompile(`(^|\s)@\S+`)
	dingtalkAmountRe   = regexp.MustCompile(`[^0-9.,-]+`)
	dingtalkWhitespace = regexp.MustCompile(`\s+`)
)

type DingtalkBotHandler struct {
	cfg          *config.Config
	adminService service.AdminService
	userService  *service.UserService
}

func NewDingtalkBotHandler(
	cfg *config.Config,
	adminService service.AdminService,
	userService *service.UserService,
) *DingtalkBotHandler {
	return &DingtalkBotHandler{
		cfg:          cfg,
		adminService: adminService,
		userService:  userService,
	}
}

type dingtalkBotRequest struct {
	Text           dingtalkBotText     `json:"text"`
	Markdown       dingtalkBotMarkdown `json:"markdown"`
	AtUsers        []dingtalkAtUser    `json:"atUsers"`
	SenderNick     string              `json:"senderNick"`
	SenderID       string              `json:"senderId"`
	SenderStaffID  string              `json:"senderStaffId"`
	ConversationID string              `json:"conversationId"`
}

type dingtalkBotText struct {
	Content string `json:"content"`
}

type dingtalkBotMarkdown struct {
	Text string `json:"text"`
}

type dingtalkAtUser struct {
	StaffID    string `json:"staffId"`
	DingtalkID string `json:"dingtalkId"`
}

func (h *DingtalkBotHandler) RechargeStatus(c *gin.Context) {
	botConfig := h.getBotConfig()
	if botConfig == nil || !botConfig.Enabled {
		c.JSON(503, createDingtalkMarkdownResponse("功能未启用", "### ⚠️ 钉钉机器人充值功能未启用"))
		return
	}
	c.JSON(200, createDingtalkMarkdownResponse("接口可用", "### ✅ 接口可用"))
}

func (h *DingtalkBotHandler) Recharge(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("dingtalk bot request read failed: %v", err)
	} else {
		logDingtalkRequest(c, bodyBytes)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	botConfig := h.getBotConfig()
	if botConfig == nil || !botConfig.Enabled {
		c.JSON(503, createDingtalkMarkdownResponse("功能未启用", "### ⚠️ 钉钉机器人充值功能未启用"))
		return
	}

	if botConfig.AccessToken != "" {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("x-dingtalk-token")
		}
		if token != botConfig.AccessToken {
			log.Printf("dingtalk bot rejected: invalid access token")
			c.JSON(401, createDingtalkMarkdownResponse("认证失败", "### ❌ Token 无效"))
			return
		}
	}

	if botConfig.SignSecret != "" {
		timestamp := c.Query("timestamp")
		sign := c.Query("sign")
		if !verifyDingtalkSignature(botConfig.SignSecret, timestamp, sign) {
			log.Printf("dingtalk bot rejected: invalid signature")
			c.JSON(401, createDingtalkMarkdownResponse("认证失败", "### ❌ 签名校验失败"))
			return
		}
	}

	var req dingtalkBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, createDingtalkMarkdownResponse("请求错误", "### ❌ 请求体解析失败"))
		return
	}

	rawContent := strings.TrimSpace(req.Text.Content)
	if rawContent == "" {
		rawContent = strings.TrimSpace(req.Markdown.Text)
	}
	content := cleanDingtalkContent(rawContent, req.AtUsers)
	if content == "" {
		c.JSON(400, dingtalkHelpResponse())
		return
	}

	senderIdentifier := strings.TrimSpace(req.SenderStaffID)
	if senderIdentifier == "" {
		senderIdentifier = strings.TrimSpace(req.SenderID)
	}

	if isDingtalkHelpCommand(content) {
		c.JSON(200, dingtalkHelpResponse())
		return
	}

	email, delta, err := parseBalanceCommand(content)
	if err != nil {
		c.JSON(400, dingtalkHelpResponse())
		return
	}

	allowedSenders := parseCommaSeparated(botConfig.AllowedSenderIDs)
	if len(allowedSenders) > 0 && !containsString(allowedSenders, senderIdentifier) {
		log.Printf("dingtalk bot rejected: sender not allowed: %s", senderIdentifier)
		c.JSON(403, createDingtalkMarkdownResponse("权限不足", "### ❌ 当前钉钉账号无权执行余额操作"))
		return
	}

	if h.userService == nil || h.adminService == nil {
		c.JSON(500, createDingtalkMarkdownResponse("服务未就绪", "### ❌ 服务未就绪，请稍后再试"))
		return
	}

	user, err := h.userService.GetByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(404, createDingtalkMarkdownResponse("用户不存在", fmt.Sprintf("### ❌ 未找到用户 %s", email)))
			return
		}
		log.Printf("dingtalk bot error: get user failed: %v", err)
		c.JSON(500, createDingtalkMarkdownResponse("服务异常", "### ❌ 服务器处理失败，请稍后再试"))
		return
	}

	oldBalance := user.Balance

	operatorName := strings.TrimSpace(req.SenderNick)
	if operatorName == "" {
		operatorName = senderIdentifier
	}
	if operatorName == "" {
		operatorName = "dingtalk-bot"
	}

	remark := buildDingtalkRemark(botConfig.DefaultRemark, operatorName, req.ConversationID)
	operation := "add"
	amount := delta
	if delta < 0 {
		operation = "subtract"
		amount = -delta
	}

	updated, err := h.adminService.UpdateUserBalance(c.Request.Context(), user.ID, amount, operation, remark)
	if err != nil {
		log.Printf("dingtalk bot error: balance update failed: %v", err)
		c.JSON(500, createDingtalkMarkdownResponse("操作失败", "### ❌ 服务器处理失败，请稍后再试"))
		return
	}

	log.Printf("dingtalk bot balance update success: %s %+0.2f by %s", user.Email, delta, operatorName)

	now := time.Now().Format("2006-01-02 15:04:05")
	c.JSON(200, createDingtalkMarkdownResponse(
		"余额调整成功",
		fmt.Sprintf("### ✅ 余额调整成功\n\n- **用户**: %s\n- **原余额**: $%.2f\n- **变动额度**: %s\n- **现余额**: $%.2f\n- **操作人**: %s\n- **操作时间**: %s",
			formatDingtalkUserDisplay(user),
			oldBalance,
			formatSignedAmount(delta),
			updated.Balance,
			operatorName,
			now,
		),
	))
}

func (h *DingtalkBotHandler) getBotConfig() *config.DingtalkBotConfig {
	if h == nil || h.cfg == nil {
		return nil
	}
	return &h.cfg.DingtalkBot
}

func createDingtalkMarkdownResponse(title, text string) gin.H {
	return gin.H{
		"msgtype": "markdown",
		"markdown": gin.H{
			"title": title,
			"text":  text,
		},
	}
}

func dingtalkHelpResponse() gin.H {
	return createDingtalkMarkdownResponse(
		"帮助",
		"### 🤖 钉钉机器人操作说明\n\n"+
			"- **help**：查看功能列表\n"+
			"- **balance <email> +10/-10**：调整用户余额\n\n"+
			"**示例**\n"+
			"- balance user@example.com +10\n"+
			"- balance user@example.com -10",
	)
}

func isDingtalkHelpCommand(content string) bool {
	content = strings.TrimSpace(strings.ToLower(content))
	return content == "help" || content == "帮助"
}

func cleanDingtalkContent(raw string, atUsers []dingtalkAtUser) string {
	content := strings.ReplaceAll(raw, "\r\n", " ")
	content = strings.ReplaceAll(content, "\n", " ")
	for _, user := range atUsers {
		if user.StaffID != "" {
			content = strings.ReplaceAll(content, "@"+user.StaffID, " ")
		}
		if user.DingtalkID != "" {
			content = strings.ReplaceAll(content, "@"+user.DingtalkID, " ")
		}
	}
	content = dingtalkAtUserRe.ReplaceAllString(content, "$1")
	content = dingtalkWhitespace.ReplaceAllString(content, " ")
	return strings.TrimSpace(content)
}

func parseBalanceCommand(content string) (string, float64, error) {
	parts := strings.Fields(content)
	if len(parts) < 3 {
		return "", 0, fmt.Errorf("invalid command")
	}
	if strings.ToLower(parts[0]) != "balance" {
		return "", 0, fmt.Errorf("invalid command")
	}
	email := strings.TrimSpace(parts[1])
	if email == "" {
		return "", 0, fmt.Errorf("missing email")
	}
	delta, err := parseSignedAmount(parts[2])
	if err != nil {
		return "", 0, err
	}
	if delta == 0 {
		return "", 0, fmt.Errorf("zero delta")
	}
	return email, delta, nil
}

func parseSignedAmount(amountText string) (float64, error) {
	trimmed := strings.TrimSpace(amountText)
	if trimmed == "" {
		return 0, fmt.Errorf("empty amount")
	}
	sign := trimmed[0]
	if sign != '+' && sign != '-' {
		return 0, fmt.Errorf("amount must include sign")
	}
	sanitized := dingtalkAmountRe.ReplaceAllString(trimmed, "")
	sanitized = strings.ReplaceAll(sanitized, ",", "")
	value, err := strconv.ParseFloat(sanitized, 64)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		value = -value
	}
	if sign == '-' {
		value = -value
	}
	return value, nil
}

func verifyDingtalkSignature(secret, timestamp, providedSign string) bool {
	if secret == "" {
		return true
	}
	if timestamp == "" || providedSign == "" {
		return false
	}
	tsMillis, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	now := time.Now().UnixMilli()
	diff := time.Duration(now-tsMillis) * time.Millisecond
	if diff < 0 {
		diff = -diff
	}
	if diff > dingtalkMaxTimeDiff {
		log.Printf("dingtalk bot rejected: timestamp expired (diff=%s)", diff)
		return false
	}
	expected := dingtalkSign(timestamp, secret)
	if decoded, err := url.QueryUnescape(providedSign); err == nil && decoded != "" {
		if hmac.Equal([]byte(expected), []byte(decoded)) {
			return true
		}
	}
	return hmac.Equal([]byte(expected), []byte(providedSign))
}

func dingtalkSign(timestamp, secret string) string {
	stringToSign := fmt.Sprintf("%s\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sum := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}

func buildDingtalkRemark(defaultRemark, operator, conversationID string) string {
	remarkParts := make([]string, 0, 3)
	if strings.TrimSpace(defaultRemark) != "" {
		remarkParts = append(remarkParts, strings.TrimSpace(defaultRemark))
	}
	if strings.TrimSpace(operator) != "" {
		remarkParts = append(remarkParts, fmt.Sprintf("From:%s", operator))
	}
	if strings.TrimSpace(conversationID) != "" {
		remarkParts = append(remarkParts, fmt.Sprintf("Conv:%s", conversationID))
	}
	return strings.Join(remarkParts, " / ")
}

func logDingtalkRequest(c *gin.Context, body []byte) {
	if c == nil || c.Request == nil {
		return
	}
	query := c.Request.URL.Query()
	queryToken := maskDingtalkSecret(query.Get("token"))
	querySign := maskDingtalkSecret(query.Get("sign"))
	headerToken := maskDingtalkSecret(c.GetHeader("x-dingtalk-token"))
	bodyText := strings.TrimSpace(string(body))
	log.Printf(
		"dingtalk bot request: method=%s path=%s token=%s timestamp=%s sign=%s header_token=%s body=%s",
		c.Request.Method,
		c.Request.URL.Path,
		queryToken,
		query.Get("timestamp"),
		querySign,
		headerToken,
		bodyText,
	)
}

func maskDingtalkSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

func parseCommaSeparated(value string) []string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func formatDingtalkUserDisplay(user *service.User) string {
	if user == nil {
		return "-"
	}
	if user.Username != "" && user.Email != "" {
		return fmt.Sprintf("%s (%s)", user.Username, user.Email)
	}
	if user.Email != "" {
		return user.Email
	}
	return user.Username
}

func formatSignedAmount(amount float64) string {
	if amount >= 0 {
		return fmt.Sprintf("+$%.2f", amount)
	}
	return fmt.Sprintf("-$%.2f", -amount)
}
