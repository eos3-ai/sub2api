package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/google/wire"
)

// ProvideAdminHandlers creates the AdminHandlers struct
func ProvideAdminHandlers(
	dashboardHandler *admin.DashboardHandler,
	userHandler *admin.UserHandler,
	groupHandler *admin.GroupHandler,
	accountHandler *admin.AccountHandler,
	paymentOrdersHandler *admin.PaymentOrdersHandler,
	invoiceHandler *admin.InvoiceHandler,
	announcementHandler *admin.AnnouncementHandler,
	dataManagementHandler *admin.DataManagementHandler,
	backupHandler *admin.BackupHandler,
	oauthHandler *admin.OAuthHandler,
	openaiOAuthHandler *admin.OpenAIOAuthHandler,
	geminiOAuthHandler *admin.GeminiOAuthHandler,
	antigravityOAuthHandler *admin.AntigravityOAuthHandler,
	proxyHandler *admin.ProxyHandler,
	redeemHandler *admin.RedeemHandler,
	promoHandler *admin.PromoHandler,
	settingHandler *admin.SettingHandler,
	opsHandler *admin.OpsHandler,
	systemHandler *admin.SystemHandler,
	subscriptionHandler *admin.SubscriptionHandler,
	usageHandler *admin.UsageHandler,
	userAttributeHandler *admin.UserAttributeHandler,
	errorPassthroughHandler *admin.ErrorPassthroughHandler,
	apiKeyHandler *admin.AdminAPIKeyHandler,
	scheduledTestHandler *admin.ScheduledTestHandler,
) *AdminHandlers {
	return &AdminHandlers{
		Dashboard:        dashboardHandler,
		User:             userHandler,
		Group:            groupHandler,
		Account:          accountHandler,
		PaymentOrders:    paymentOrdersHandler,
		Invoices:         invoiceHandler,
		Announcement:     announcementHandler,
		DataManagement:   dataManagementHandler,
		Backup:           backupHandler,
		OAuth:            oauthHandler,
		OpenAIOAuth:      openaiOAuthHandler,
		GeminiOAuth:      geminiOAuthHandler,
		AntigravityOAuth: antigravityOAuthHandler,
		Proxy:            proxyHandler,
		Redeem:           redeemHandler,
		Promo:            promoHandler,
		Setting:          settingHandler,
		Ops:              opsHandler,
		System:           systemHandler,
		Subscription:     subscriptionHandler,
		Usage:            usageHandler,
		UserAttribute:    userAttributeHandler,
		ErrorPassthrough: errorPassthroughHandler,
		APIKey:           apiKeyHandler,
		ScheduledTest:    scheduledTestHandler,
	}
}

// ProvideSystemHandler creates admin.SystemHandler with UpdateService
func ProvideSystemHandler(updateService *service.UpdateService, lockService *service.SystemOperationLockService) *admin.SystemHandler {
	return admin.NewSystemHandler(updateService, lockService)
}

// ProvideSettingHandler creates SettingHandler with version from BuildInfo
func ProvideSettingHandler(settingService *service.SettingService, buildInfo BuildInfo) *SettingHandler {
	return NewSettingHandler(settingService, buildInfo.Version)
}

// ProvideHandlers creates the Handlers struct
func ProvideHandlers(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	apiKeyHandler *APIKeyHandler,
	usageHandler *UsageHandler,
	redeemHandler *RedeemHandler,
	subscriptionHandler *SubscriptionHandler,
	paymentHandler *PaymentHandler,
	invoiceHandler *InvoiceHandler,
	promotionHandler *PromotionHandler,
	referralHandler *ReferralHandler,
	announcementHandler *AnnouncementHandler,
	adminHandlers *AdminHandlers,
	gatewayHandler *GatewayHandler,
	openaiGatewayHandler *OpenAIGatewayHandler,
	soraGatewayHandler *SoraGatewayHandler,
	soraClientHandler *SoraClientHandler,
	settingHandler *SettingHandler,
	dingtalkBotHandler *DingtalkBotHandler,
	totpHandler *TotpHandler,
	_ *service.IdempotencyCoordinator,
	_ *service.IdempotencyCleanupService,
) *Handlers {
	return &Handlers{
		Auth:          authHandler,
		User:          userHandler,
		APIKey:        apiKeyHandler,
		Usage:         usageHandler,
		Redeem:        redeemHandler,
		Subscription:  subscriptionHandler,
		Payment:       paymentHandler,
		Invoice:       invoiceHandler,
		Promotion:     promotionHandler,
		Referral:      referralHandler,
		Announcement:  announcementHandler,
		Admin:         adminHandlers,
		Gateway:       gatewayHandler,
		OpenAIGateway: openaiGatewayHandler,
		SoraGateway:   soraGatewayHandler,
		SoraClient:    soraClientHandler,
		Setting:       settingHandler,
		DingtalkBot:   dingtalkBotHandler,
		Totp:          totpHandler,
	}
}

// ProviderSet is the Wire provider set for all handlers
var ProviderSet = wire.NewSet(
	// Top-level handlers
	NewAuthHandler,
	NewUserHandler,
	NewAPIKeyHandler,
	NewUsageHandler,
	NewRedeemHandler,
	NewSubscriptionHandler,
	NewPaymentHandler,
	NewInvoiceHandler,
	NewPromotionHandler,
	NewReferralHandler,
	NewAnnouncementHandler,
	NewGatewayHandler,
	NewOpenAIGatewayHandler,
	NewDingtalkBotHandler,
	NewSoraGatewayHandler,
	NewTotpHandler,
	ProvideSettingHandler,

	// Admin handlers
	admin.NewDashboardHandler,
	admin.NewUserHandler,
	admin.NewGroupHandler,
	admin.NewAccountHandler,
	admin.NewPaymentOrdersHandler,
	admin.NewInvoiceHandler,
	admin.NewAnnouncementHandler,
	admin.NewDataManagementHandler,
	admin.NewBackupHandler,
	admin.NewOAuthHandler,
	admin.NewOpenAIOAuthHandler,
	admin.NewGeminiOAuthHandler,
	admin.NewAntigravityOAuthHandler,
	admin.NewProxyHandler,
	admin.NewRedeemHandler,
	admin.NewPromoHandler,
	admin.NewSettingHandler,
	admin.NewOpsHandler,
	ProvideSystemHandler,
	admin.NewSubscriptionHandler,
	admin.NewUsageHandler,
	admin.NewUserAttributeHandler,
	admin.NewErrorPassthroughHandler,
	admin.NewAdminAPIKeyHandler,
	admin.NewScheduledTestHandler,

	// AdminHandlers and Handlers constructors
	ProvideAdminHandlers,
	ProvideHandlers,
)
