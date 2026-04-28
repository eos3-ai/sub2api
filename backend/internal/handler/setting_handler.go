package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// OcpcEventRequest 百度 OCPC 转化事件上报请求
type OcpcEventRequest struct {
	BdVid      string `json:"bd_vid" binding:"required"`
	LandingUrl string `json:"landing_url"`
	NewType    int    `json:"new_type"`
}

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService *service.SettingService
	version        string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// ReportOcpcEvent 上报百度 OCPC 转化事件（公开接口，仅允许关键页面浏览事件）
// POST /api/v1/public/ocpc
func (h *SettingHandler) ReportOcpcEvent(c *gin.Context) {
	var req OcpcEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	landingNewType := service.GetBaiduOcpcLandingNewType()
	// 仅允许上报关键页面浏览事件，防止公开接口被滥用
	if req.NewType != landingNewType {
		response.BadRequest(c, "Invalid new_type: only configured landing new_type is allowed via public API")
		return
	}
	service.ReportBaiduOcpcEvent(req.BdVid, req.LandingUrl, req.NewType)
	response.Success(c, gin.H{"ok": true})
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PublicSettings{
		RegistrationEnabled:         settings.RegistrationEnabled,
		EmailVerifyEnabled:          settings.EmailVerifyEnabled,
		PromoCodeEnabled:            settings.PromoCodeEnabled,
		PasswordResetEnabled:        settings.PasswordResetEnabled,
		InvitationCodeEnabled:       settings.InvitationCodeEnabled,
		TotpEnabled:                 settings.TotpEnabled,
		TurnstileEnabled:            settings.TurnstileEnabled,
		TurnstileSiteKey:            settings.TurnstileSiteKey,
		SiteName:                    settings.SiteName,
		SiteLogo:                    settings.SiteLogo,
		SiteSubtitle:                settings.SiteSubtitle,
		APIBaseURL:                  settings.APIBaseURL,
		ContactInfo:                 settings.ContactInfo,
		DocURL:                      settings.DocURL,
		HomeContent:                 settings.HomeContent,
		HideCcsImportButton:         settings.HideCcsImportButton,
		PaymentEnabled:              settings.PaymentEnabled,
		PurchaseSubscriptionEnabled: settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:     settings.PurchaseSubscriptionURL,
		ClaudeOfficialURL:           settings.ClaudeOfficialURL,
		CodexOfficialURL:            settings.CodexOfficialURL,
		GeminiOfficialURL:           settings.GeminiOfficialURL,
		LinuxDoOAuthEnabled:         settings.LinuxDoOAuthEnabled,
		Version:                     h.version,
		BaiduTongjiID:               settings.BaiduTongjiID,
		BaiduOcpcLandingNewType:     settings.BaiduOcpcLandingNewType,
	})
}
