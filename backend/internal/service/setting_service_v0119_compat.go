package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

var (
	ErrDefaultSubGroupInvalid = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_INVALID",
		"default subscription group must be an existing subscription group",
	)
	ErrDefaultSubGroupDuplicate = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_DUPLICATE",
		"default subscription groups must not contain duplicates",
	)
)

const (
	SettingKeyRegistrationEmailSuffixWhitelist         = "registration_email_suffix_whitelist"
	SettingKeyAffiliateEnabled                         = "affiliate_enabled"
	SettingKeyAffiliateRebateRate                      = "affiliate_rebate_rate"
	SettingKeyAffiliateRebateFreezeHours               = "affiliate_rebate_freeze_hours"
	SettingKeyAffiliateRebateDurationDays              = "affiliate_rebate_duration_days"
	SettingKeyAffiliateRebatePerInviteeCap             = "affiliate_rebate_per_invitee_cap"
	SettingKeyTableDefaultPageSize                     = "table_default_page_size"
	SettingKeyTablePageSizeOptions                     = "table_page_size_options"
	SettingKeyCustomMenuItems                          = "custom_menu_items"
	SettingKeyCustomEndpoints                          = "custom_endpoints"
	SettingKeyDefaultSubscriptions                     = "default_subscriptions"
	SettingKeyDefaultUserRPMLimit                      = "default_user_rpm_limit"
	SettingKeyAuthSourceDefaultEmailBalance            = "auth_source_default_email_balance"
	SettingKeyAuthSourceDefaultEmailConcurrency        = "auth_source_default_email_concurrency"
	SettingKeyAuthSourceDefaultEmailSubscriptions      = "auth_source_default_email_subscriptions"
	SettingKeyAuthSourceDefaultEmailGrantOnSignup      = "auth_source_default_email_grant_on_signup"
	SettingKeyAuthSourceDefaultEmailGrantOnFirstBind   = "auth_source_default_email_grant_on_first_bind"
	SettingKeyAuthSourceDefaultLinuxDoBalance          = "auth_source_default_linuxdo_balance"
	SettingKeyAuthSourceDefaultLinuxDoConcurrency      = "auth_source_default_linuxdo_concurrency"
	SettingKeyAuthSourceDefaultLinuxDoSubscriptions    = "auth_source_default_linuxdo_subscriptions"
	SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup    = "auth_source_default_linuxdo_grant_on_signup"
	SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind = "auth_source_default_linuxdo_grant_on_first_bind"
	SettingKeyAuthSourceDefaultOIDCBalance             = "auth_source_default_oidc_balance"
	SettingKeyAuthSourceDefaultOIDCConcurrency         = "auth_source_default_oidc_concurrency"
	SettingKeyAuthSourceDefaultOIDCSubscriptions       = "auth_source_default_oidc_subscriptions"
	SettingKeyAuthSourceDefaultOIDCGrantOnSignup       = "auth_source_default_oidc_grant_on_signup"
	SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind    = "auth_source_default_oidc_grant_on_first_bind"
	SettingKeyAuthSourceDefaultWeChatBalance           = "auth_source_default_wechat_balance"
	SettingKeyAuthSourceDefaultWeChatConcurrency       = "auth_source_default_wechat_concurrency"
	SettingKeyAuthSourceDefaultWeChatSubscriptions     = "auth_source_default_wechat_subscriptions"
	SettingKeyAuthSourceDefaultWeChatGrantOnSignup     = "auth_source_default_wechat_grant_on_signup"
	SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind  = "auth_source_default_wechat_grant_on_first_bind"
	SettingKeyForceEmailOnThirdPartySignup             = "force_email_on_third_party_signup"
)

// cachedVersionBounds 缓存 Claude Code 版本号上下限（进程内缓存，60s TTL）
type cachedVersionBounds struct {
	min       string
	max       string
	expiresAt int64
}

var versionBoundsCache atomic.Value // *cachedVersionBounds
var versionBoundsSF singleflight.Group

const versionBoundsCacheTTL = 60 * time.Second
const versionBoundsErrorTTL = 5 * time.Second
const versionBoundsDBTimeout = 5 * time.Second

// cachedBackendMode Backend Mode cache (in-process, 60s TTL)
type cachedBackendMode struct {
	value     bool
	expiresAt int64
}

var backendModeCache atomic.Value // *cachedBackendMode
var backendModeSF singleflight.Group

const backendModeCacheTTL = 60 * time.Second
const backendModeErrorTTL = 5 * time.Second
const backendModeDBTimeout = 5 * time.Second

// cachedGatewayForwardingSettings 缓存网关转发行为设置（进程内缓存，60s TTL）
type cachedGatewayForwardingSettings struct {
	fingerprintUnification bool
	metadataPassthrough    bool
	cchSigning             bool
	expiresAt              int64
}

var gatewayForwardingCache atomic.Value // *cachedGatewayForwardingSettings
var gatewayForwardingSF singleflight.Group

const gatewayForwardingCacheTTL = 60 * time.Second
const gatewayForwardingErrorTTL = 5 * time.Second
const gatewayForwardingDBTimeout = 5 * time.Second

// DefaultSubscriptionGroupReader validates group references used by default subscriptions.
type DefaultSubscriptionGroupReader interface {
	GetByID(ctx context.Context, id int64) (*Group, error)
}

type ProviderDefaultGrantSettings struct {
	Balance          float64
	Concurrency      int
	Subscriptions    []DefaultSubscriptionSetting
	GrantOnSignup    bool
	GrantOnFirstBind bool
}

type AuthSourceDefaultSettings struct {
	Email                        ProviderDefaultGrantSettings
	LinuxDo                      ProviderDefaultGrantSettings
	OIDC                         ProviderDefaultGrantSettings
	WeChat                       ProviderDefaultGrantSettings
	ForceEmailOnThirdPartySignup bool
}

type authSourceDefaultKeySet struct {
	balance          string
	concurrency      string
	subscriptions    string
	grantOnSignup    string
	grantOnFirstBind string
}

var (
	emailAuthSourceDefaultKeys = authSourceDefaultKeySet{
		balance:          SettingKeyAuthSourceDefaultEmailBalance,
		concurrency:      SettingKeyAuthSourceDefaultEmailConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultEmailSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultEmailGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultEmailGrantOnFirstBind,
	}
	linuxDoAuthSourceDefaultKeys = authSourceDefaultKeySet{
		balance:          SettingKeyAuthSourceDefaultLinuxDoBalance,
		concurrency:      SettingKeyAuthSourceDefaultLinuxDoConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultLinuxDoSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind,
	}
	oidcAuthSourceDefaultKeys = authSourceDefaultKeySet{
		balance:          SettingKeyAuthSourceDefaultOIDCBalance,
		concurrency:      SettingKeyAuthSourceDefaultOIDCConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultOIDCSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultOIDCGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind,
	}
	weChatAuthSourceDefaultKeys = authSourceDefaultKeySet{
		balance:          SettingKeyAuthSourceDefaultWeChatBalance,
		concurrency:      SettingKeyAuthSourceDefaultWeChatConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultWeChatSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultWeChatGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind,
	}
)

const (
	defaultAuthSourceBalance     = 0
	defaultAuthSourceConcurrency = 5
	defaultWeChatConnectMode     = "open"
	defaultWeChatConnectScopes   = "snsapi_login"
	defaultWeChatConnectFrontend = "/auth/wechat/callback"
)

func (s *SettingService) SetDefaultSubscriptionGroupReader(reader DefaultSubscriptionGroupReader) {
	s.defaultSubGroupReader = reader
}

func (s *SettingService) SetProxyRepository(repo ProxyRepository) {
	s.proxyRepo = repo
}

func normalizeWeChatConnectModeSetting(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mp":
		return "mp"
	case "mobile":
		return "mobile"
	default:
		return "open"
	}
}

func defaultWeChatConnectScopeForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return "snsapi_userinfo"
	case "mobile":
		return ""
	default:
		return defaultWeChatConnectScopes
	}
}

func normalizeWeChatConnectScopeSetting(raw, mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		switch strings.TrimSpace(raw) {
		case "snsapi_base":
			return "snsapi_base"
		case "snsapi_userinfo":
			return "snsapi_userinfo"
		default:
			return defaultWeChatConnectScopeForMode(mode)
		}
	case "mobile":
		return ""
	default:
		return defaultWeChatConnectScopes
	}
}

func parseWeChatConnectCapabilitySettings(settings map[string]string, enabled bool, mode string) (bool, bool, bool) {
	mode = normalizeWeChatConnectModeSetting(mode)
	rawOpen, hasOpen := settings[SettingKeyWeChatConnectOpenEnabled]
	rawMP, hasMP := settings[SettingKeyWeChatConnectMPEnabled]
	rawMobile, hasMobile := settings[SettingKeyWeChatConnectMobileEnabled]
	openConfigured := hasOpen && strings.TrimSpace(rawOpen) != ""
	mpConfigured := hasMP && strings.TrimSpace(rawMP) != ""
	mobileConfigured := hasMobile && strings.TrimSpace(rawMobile) != ""

	if openConfigured || mpConfigured || mobileConfigured {
		return strings.TrimSpace(rawOpen) == "true", strings.TrimSpace(rawMP) == "true", strings.TrimSpace(rawMobile) == "true"
	}
	if !enabled {
		return false, false, false
	}
	switch mode {
	case "mp":
		return false, true, false
	case "mobile":
		return false, false, true
	default:
		return true, false, false
	}
}

func normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled bool, mode string) string {
	mode = normalizeWeChatConnectModeSetting(mode)
	switch mode {
	case "open":
		if openEnabled {
			return "open"
		}
	case "mp":
		if mpEnabled {
			return "mp"
		}
	case "mobile":
		if mobileEnabled {
			return "mobile"
		}
	}
	switch {
	case openEnabled:
		return "open"
	case mpEnabled:
		return "mp"
	case mobileEnabled:
		return "mobile"
	default:
		return mode
	}
}

func mergeWeChatConnectCapabilitySettings(settings map[string]string, base config.WeChatConnectConfig, enabled bool, mode string) (bool, bool, bool) {
	mode = normalizeWeChatConnectModeSetting(firstNonEmpty(mode, base.Mode))
	rawOpen, hasOpen := settings[SettingKeyWeChatConnectOpenEnabled]
	rawMP, hasMP := settings[SettingKeyWeChatConnectMPEnabled]
	rawMobile, hasMobile := settings[SettingKeyWeChatConnectMobileEnabled]
	openConfigured := hasOpen && strings.TrimSpace(rawOpen) != ""
	mpConfigured := hasMP && strings.TrimSpace(rawMP) != ""
	mobileConfigured := hasMobile && strings.TrimSpace(rawMobile) != ""

	if openConfigured || mpConfigured || mobileConfigured {
		openEnabled := strings.TrimSpace(rawOpen) == "true"
		mpEnabled := strings.TrimSpace(rawMP) == "true"
		mobileEnabled := strings.TrimSpace(rawMobile) == "true"
		_, enabledConfigured := settings[SettingKeyWeChatConnectEnabled]
		if !enabledConfigured &&
			enabled &&
			!openEnabled &&
			!mpEnabled &&
			!mobileEnabled &&
			(base.OpenEnabled || base.MPEnabled || base.MobileEnabled) {
			return base.OpenEnabled, base.MPEnabled, base.MobileEnabled
		}
		return openEnabled, mpEnabled, mobileEnabled
	}
	if !enabled {
		return false, false, false
	}
	if base.OpenEnabled || base.MPEnabled || base.MobileEnabled {
		return base.OpenEnabled, base.MPEnabled, base.MobileEnabled
	}
	return parseWeChatConnectCapabilitySettings(settings, enabled, mode)
}

func (s *SettingService) effectiveWeChatConnectOAuthConfig(settings map[string]string) WeChatConnectOAuthConfig {
	base := config.WeChatConnectConfig{}
	if s != nil && s.cfg != nil {
		base = s.cfg.WeChat
	}

	enabled := base.Enabled
	if raw, ok := settings[SettingKeyWeChatConnectEnabled]; ok {
		enabled = strings.TrimSpace(raw) == "true"
	}

	legacyAppID := strings.TrimSpace(firstNonEmpty(
		settings[SettingKeyWeChatConnectAppID],
		base.AppID,
		base.OpenAppID,
		base.MPAppID,
		base.MobileAppID,
	))
	legacyAppSecret := strings.TrimSpace(firstNonEmpty(
		settings[SettingKeyWeChatConnectAppSecret],
		base.AppSecret,
		base.OpenAppSecret,
		base.MPAppSecret,
		base.MobileAppSecret,
	))
	openAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppID], base.OpenAppID, legacyAppID))
	openAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppSecret], base.OpenAppSecret, legacyAppSecret))
	mpAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppID], base.MPAppID, legacyAppID))
	mpAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppSecret], base.MPAppSecret, legacyAppSecret))
	mobileAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppID], base.MobileAppID, legacyAppID))
	mobileAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppSecret], base.MobileAppSecret, legacyAppSecret))

	modeRaw := firstNonEmpty(settings[SettingKeyWeChatConnectMode], base.Mode)
	openEnabled, mpEnabled, mobileEnabled := mergeWeChatConnectCapabilitySettings(settings, base, enabled, modeRaw)
	mode := normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled, modeRaw)

	return WeChatConnectOAuthConfig{
		Enabled:             enabled,
		LegacyAppID:         legacyAppID,
		LegacyAppSecret:     legacyAppSecret,
		OpenAppID:           openAppID,
		OpenAppSecret:       openAppSecret,
		MPAppID:             mpAppID,
		MPAppSecret:         mpAppSecret,
		MobileAppID:         mobileAppID,
		MobileAppSecret:     mobileAppSecret,
		OpenEnabled:         openEnabled,
		MPEnabled:           mpEnabled,
		MobileEnabled:       mobileEnabled,
		Mode:                mode,
		Scopes:              normalizeWeChatConnectScopeSetting(firstNonEmpty(settings[SettingKeyWeChatConnectScopes], base.Scopes), mode),
		RedirectURL:         strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectRedirectURL], base.RedirectURL)),
		FrontendRedirectURL: strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectFrontendRedirectURL], base.FrontendRedirectURL, defaultWeChatConnectFrontend)),
	}
}

func (s *SettingService) weChatOAuthCapabilitiesFromSettings(settings map[string]string) (bool, bool, bool, bool) {
	cfg := s.effectiveWeChatConnectOAuthConfig(settings)
	return cfg.Enabled, cfg.OpenEnabled, cfg.MPEnabled, cfg.MobileEnabled
}

func (s *SettingService) parseWeChatConnectOAuthConfig(settings map[string]string) (WeChatConnectOAuthConfig, error) {
	cfg := s.effectiveWeChatConnectOAuthConfig(settings)
	if !cfg.Enabled || (!cfg.OpenEnabled && !cfg.MPEnabled && !cfg.MobileEnabled) {
		return WeChatConnectOAuthConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "wechat oauth is disabled")
	}
	if v := strings.TrimSpace(cfg.RedirectURL); v != "" {
		if err := config.ValidateAbsoluteHTTPURL(v); err != nil {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth redirect url invalid")
		}
	}
	if err := config.ValidateFrontendRedirectURL(cfg.FrontendRedirectURL); err != nil {
		return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth frontend redirect url invalid")
	}
	return cfg, nil
}

func oidcUsePKCECompatibilityDefault(base config.OIDCConnectConfig) bool {
	if base.UsePKCEExplicit {
		return base.UsePKCE
	}
	return true
}

func oidcValidateIDTokenCompatibilityDefault(base config.OIDCConnectConfig) bool {
	if base.ValidateIDTokenExplicit {
		return base.ValidateIDToken
	}
	return true
}

func (s *SettingService) GetRegistrationEmailSuffixWhitelist(ctx context.Context) []string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationEmailSuffixWhitelist)
	if err != nil {
		return []string{}
	}
	return ParseRegistrationEmailSuffixWhitelist(value)
}

func (s *SettingService) GetDefaultUserRPMLimit(ctx context.Context) int {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultUserRPMLimit)
	if err != nil || value == "" {
		return 0
	}
	if v, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && v >= 0 {
		return v
	}
	return 0
}

func (s *SettingService) GetDefaultSubscriptions(ctx context.Context) []DefaultSubscriptionSetting {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultSubscriptions)
	if err != nil {
		return nil
	}
	return parseDefaultSubscriptions(value)
}

func (s *SettingService) GetAuthSourceDefaultSettings(ctx context.Context) (*AuthSourceDefaultSettings, error) {
	keys := []string{
		SettingKeyAuthSourceDefaultEmailBalance,
		SettingKeyAuthSourceDefaultEmailConcurrency,
		SettingKeyAuthSourceDefaultEmailSubscriptions,
		SettingKeyAuthSourceDefaultEmailGrantOnSignup,
		SettingKeyAuthSourceDefaultEmailGrantOnFirstBind,
		SettingKeyAuthSourceDefaultLinuxDoBalance,
		SettingKeyAuthSourceDefaultLinuxDoConcurrency,
		SettingKeyAuthSourceDefaultLinuxDoSubscriptions,
		SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup,
		SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind,
		SettingKeyAuthSourceDefaultOIDCBalance,
		SettingKeyAuthSourceDefaultOIDCConcurrency,
		SettingKeyAuthSourceDefaultOIDCSubscriptions,
		SettingKeyAuthSourceDefaultOIDCGrantOnSignup,
		SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind,
		SettingKeyAuthSourceDefaultWeChatBalance,
		SettingKeyAuthSourceDefaultWeChatConcurrency,
		SettingKeyAuthSourceDefaultWeChatSubscriptions,
		SettingKeyAuthSourceDefaultWeChatGrantOnSignup,
		SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind,
		SettingKeyForceEmailOnThirdPartySignup,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get auth source default settings: %w", err)
	}

	return &AuthSourceDefaultSettings{
		Email:                        parseProviderDefaultGrantSettings(settings, emailAuthSourceDefaultKeys),
		LinuxDo:                      parseProviderDefaultGrantSettings(settings, linuxDoAuthSourceDefaultKeys),
		OIDC:                         parseProviderDefaultGrantSettings(settings, oidcAuthSourceDefaultKeys),
		WeChat:                       parseProviderDefaultGrantSettings(settings, weChatAuthSourceDefaultKeys),
		ForceEmailOnThirdPartySignup: settings[SettingKeyForceEmailOnThirdPartySignup] == "true",
	}, nil
}

func authSourceSignupSettings(defaults *AuthSourceDefaultSettings, signupSource string) (ProviderDefaultGrantSettings, bool) {
	if defaults == nil {
		return ProviderDefaultGrantSettings{}, false
	}

	switch strings.ToLower(strings.TrimSpace(signupSource)) {
	case "email":
		return defaults.Email, true
	case "linuxdo":
		return defaults.LinuxDo, true
	case "oidc":
		return defaults.OIDC, true
	case "wechat":
		return defaults.WeChat, true
	default:
		return ProviderDefaultGrantSettings{}, false
	}
}

func (s *SettingService) ResolveAuthSourceGrantSettings(ctx context.Context, signupSource string, firstBind bool) (ProviderDefaultGrantSettings, bool, error) {
	result := ProviderDefaultGrantSettings{
		Balance:       s.GetDefaultBalance(ctx),
		Concurrency:   s.GetDefaultConcurrency(ctx),
		Subscriptions: s.GetDefaultSubscriptions(ctx),
	}

	defaults, err := s.GetAuthSourceDefaultSettings(ctx)
	if err != nil {
		return result, false, err
	}

	providerDefaults, ok := authSourceSignupSettings(defaults, signupSource)
	if !ok {
		return result, false, nil
	}

	enabled := providerDefaults.GrantOnSignup
	if firstBind {
		enabled = providerDefaults.GrantOnFirstBind
	}
	if !enabled {
		return result, false, nil
	}

	return mergeProviderDefaultGrantSettings(result, providerDefaults), true, nil
}

func (s *SettingService) UpdateAuthSourceDefaultSettings(ctx context.Context, settings *AuthSourceDefaultSettings) error {
	updates, err := s.buildAuthSourceDefaultUpdates(ctx, settings)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	if err := s.settingRepo.SetMultiple(ctx, updates); err != nil {
		return fmt.Errorf("update auth source default settings: %w", err)
	}
	return nil
}

func (s *SettingService) buildAuthSourceDefaultUpdates(ctx context.Context, settings *AuthSourceDefaultSettings) (map[string]string, error) {
	if settings == nil {
		return nil, nil
	}

	for _, subscriptions := range [][]DefaultSubscriptionSetting{
		settings.Email.Subscriptions,
		settings.LinuxDo.Subscriptions,
		settings.OIDC.Subscriptions,
		settings.WeChat.Subscriptions,
	} {
		if err := s.validateDefaultSubscriptionGroups(ctx, subscriptions); err != nil {
			return nil, err
		}
	}

	updates := make(map[string]string, 21)
	writeProviderDefaultGrantUpdates(updates, emailAuthSourceDefaultKeys, settings.Email)
	writeProviderDefaultGrantUpdates(updates, linuxDoAuthSourceDefaultKeys, settings.LinuxDo)
	writeProviderDefaultGrantUpdates(updates, oidcAuthSourceDefaultKeys, settings.OIDC)
	writeProviderDefaultGrantUpdates(updates, weChatAuthSourceDefaultKeys, settings.WeChat)
	updates[SettingKeyForceEmailOnThirdPartySignup] = strconv.FormatBool(settings.ForceEmailOnThirdPartySignup)
	return updates, nil
}

func (s *SettingService) validateDefaultSubscriptionGroups(ctx context.Context, items []DefaultSubscriptionSetting) error {
	if len(items) == 0 {
		return nil
	}

	checked := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item.GroupID <= 0 {
			continue
		}
		if _, ok := checked[item.GroupID]; ok {
			return ErrDefaultSubGroupDuplicate.WithMetadata(map[string]string{
				"group_id": strconv.FormatInt(item.GroupID, 10),
			})
		}
		checked[item.GroupID] = struct{}{}
		if s.defaultSubGroupReader == nil {
			continue
		}

		group, err := s.defaultSubGroupReader.GetByID(ctx, item.GroupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				return ErrDefaultSubGroupInvalid.WithMetadata(map[string]string{
					"group_id": strconv.FormatInt(item.GroupID, 10),
				})
			}
			return fmt.Errorf("get default subscription group %d: %w", item.GroupID, err)
		}
		if !group.IsSubscriptionType() {
			return ErrDefaultSubGroupInvalid.WithMetadata(map[string]string{
				"group_id": strconv.FormatInt(item.GroupID, 10),
			})
		}
	}

	return nil
}

func parseDefaultSubscriptions(raw string) []DefaultSubscriptionSetting {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var items []DefaultSubscriptionSetting
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}

	normalized := make([]DefaultSubscriptionSetting, 0, len(items))
	for _, item := range items {
		if item.GroupID <= 0 || item.ValidityDays <= 0 {
			continue
		}
		if item.ValidityDays > MaxValidityDays {
			item.ValidityDays = MaxValidityDays
		}
		normalized = append(normalized, item)
	}

	return normalized
}

func parseProviderDefaultGrantSettings(settings map[string]string, keys authSourceDefaultKeySet) ProviderDefaultGrantSettings {
	result := ProviderDefaultGrantSettings{
		Balance:          defaultAuthSourceBalance,
		Concurrency:      defaultAuthSourceConcurrency,
		Subscriptions:    []DefaultSubscriptionSetting{},
		GrantOnSignup:    false,
		GrantOnFirstBind: false,
	}

	if v, err := strconv.ParseFloat(strings.TrimSpace(settings[keys.balance]), 64); err == nil {
		result.Balance = v
	}
	if v, err := strconv.Atoi(strings.TrimSpace(settings[keys.concurrency])); err == nil {
		result.Concurrency = v
	}
	if items := parseDefaultSubscriptions(settings[keys.subscriptions]); items != nil {
		result.Subscriptions = items
	}
	if raw, ok := settings[keys.grantOnSignup]; ok {
		result.GrantOnSignup = raw == "true"
	}
	if raw, ok := settings[keys.grantOnFirstBind]; ok {
		result.GrantOnFirstBind = raw == "true"
	}

	return result
}

func writeProviderDefaultGrantUpdates(updates map[string]string, keys authSourceDefaultKeySet, settings ProviderDefaultGrantSettings) {
	updates[keys.balance] = strconv.FormatFloat(settings.Balance, 'f', 8, 64)
	updates[keys.concurrency] = strconv.Itoa(settings.Concurrency)

	subscriptions := settings.Subscriptions
	if subscriptions == nil {
		subscriptions = []DefaultSubscriptionSetting{}
	}
	raw, err := json.Marshal(subscriptions)
	if err != nil {
		raw = []byte("[]")
	}
	updates[keys.subscriptions] = string(raw)
	updates[keys.grantOnSignup] = strconv.FormatBool(settings.GrantOnSignup)
	updates[keys.grantOnFirstBind] = strconv.FormatBool(settings.GrantOnFirstBind)
}

func mergeProviderDefaultGrantSettings(globalDefaults ProviderDefaultGrantSettings, providerDefaults ProviderDefaultGrantSettings) ProviderDefaultGrantSettings {
	result := ProviderDefaultGrantSettings{
		Balance:          globalDefaults.Balance,
		Concurrency:      globalDefaults.Concurrency,
		Subscriptions:    append([]DefaultSubscriptionSetting(nil), globalDefaults.Subscriptions...),
		GrantOnSignup:    providerDefaults.GrantOnSignup,
		GrantOnFirstBind: providerDefaults.GrantOnFirstBind,
	}

	if providerDefaults.Balance != defaultAuthSourceBalance {
		result.Balance = providerDefaults.Balance
	}
	if providerDefaults.Concurrency > 0 && providerDefaults.Concurrency != defaultAuthSourceConcurrency {
		result.Concurrency = providerDefaults.Concurrency
	}
	if len(providerDefaults.Subscriptions) > 0 {
		result.Subscriptions = append([]DefaultSubscriptionSetting(nil), providerDefaults.Subscriptions...)
	}

	return result
}

func parseTablePreferences(defaultPageSizeRaw, optionsRaw string) (int, []int) {
	defaultPageSize := 20
	if v, err := strconv.Atoi(strings.TrimSpace(defaultPageSizeRaw)); err == nil {
		defaultPageSize = v
	}

	var options []int
	if strings.TrimSpace(optionsRaw) != "" {
		_ = json.Unmarshal([]byte(optionsRaw), &options)
	}

	return normalizeTablePreferences(defaultPageSize, options)
}

func normalizeTablePreferences(defaultPageSize int, options []int) (int, []int) {
	const minPageSize = 5
	const maxPageSize = 1000
	const fallbackPageSize = 20

	seen := make(map[int]struct{}, len(options))
	normalizedOptions := make([]int, 0, len(options))
	for _, option := range options {
		if option < minPageSize || option > maxPageSize {
			continue
		}
		if _, ok := seen[option]; ok {
			continue
		}
		seen[option] = struct{}{}
		normalizedOptions = append(normalizedOptions, option)
	}
	if len(normalizedOptions) == 0 {
		normalizedOptions = []int{10, 20, 50, 100}
	}

	validDefault := false
	for _, option := range normalizedOptions {
		if option == defaultPageSize {
			validDefault = true
			break
		}
	}
	if !validDefault {
		defaultPageSize = fallbackPageSize
	}

	return defaultPageSize, normalizedOptions
}

func normalizeVisibleMethodSettingSource(method, source string, enabled bool) (string, error) {
	_ = enabled
	source = strings.TrimSpace(source)
	if source == "" {
		return "", nil
	}

	normalized := NormalizeVisibleMethodSource(method, source)
	if normalized == "" {
		return "", infraerrors.BadRequest(
			"INVALID_PAYMENT_VISIBLE_METHOD_SOURCE",
			fmt.Sprintf("%s source must be one of the supported payment providers", method),
		)
	}
	return normalized, nil
}

func DefaultWeChatConnectScopesForMode(mode string) string {
	return defaultWeChatConnectScopeForMode(mode)
}

func (s *SettingService) OIDCSecurityWriteDefaults(ctx context.Context) (bool, bool, error) {
	settings, err := s.GetAllSettings(ctx)
	if err != nil {
		return false, false, err
	}
	return settings.OIDCConnectUsePKCE, settings.OIDCConnectValidateIDToken, nil
}

func (s *SettingService) UpdateSettingsWithAuthSourceDefaults(ctx context.Context, settings *SystemSettings, authSourceDefaults *AuthSourceDefaultSettings) error {
	if err := s.UpdateSettings(ctx, settings); err != nil {
		return err
	}
	return s.UpdateAuthSourceDefaultSettings(ctx, authSourceDefaults)
}

func (s *SettingService) GetClaudeCodeVersionBounds(ctx context.Context) (string, string) {
	settings, err := s.GetAllSettings(ctx)
	if err != nil {
		return "", ""
	}
	return settings.MinClaudeCodeVersion, settings.MaxClaudeCodeVersion
}

func (s *SettingService) GetOIDCConnectOAuthConfig(ctx context.Context) (config.OIDCConnectConfig, error) {
	var effective config.OIDCConnectConfig
	if s != nil && s.cfg != nil {
		effective = s.cfg.OIDC
	}

	settings, err := s.GetAllSettings(ctx)
	if err != nil {
		return config.OIDCConnectConfig{}, err
	}

	effective.Enabled = settings.OIDCConnectEnabled
	effective.ProviderName = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectProviderName, effective.ProviderName, "OIDC"))
	effective.ClientID = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectClientID, effective.ClientID))
	effective.ClientSecret = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectClientSecret, effective.ClientSecret))
	effective.IssuerURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectIssuerURL, effective.IssuerURL))
	effective.DiscoveryURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectDiscoveryURL, effective.DiscoveryURL))
	effective.AuthorizeURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectAuthorizeURL, effective.AuthorizeURL))
	effective.TokenURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectTokenURL, effective.TokenURL))
	effective.UserInfoURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectUserInfoURL, effective.UserInfoURL))
	effective.JWKSURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectJWKSURL, effective.JWKSURL))
	effective.Scopes = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectScopes, effective.Scopes, "openid email profile"))
	effective.RedirectURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectRedirectURL, effective.RedirectURL))
	effective.FrontendRedirectURL = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectFrontendRedirectURL, effective.FrontendRedirectURL, "/auth/oidc/callback"))
	effective.TokenAuthMethod = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectTokenAuthMethod, effective.TokenAuthMethod, "client_secret_post"))
	effective.UsePKCE = settings.OIDCConnectUsePKCE
	effective.ValidateIDToken = settings.OIDCConnectValidateIDToken
	effective.AllowedSigningAlgs = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectAllowedSigningAlgs, effective.AllowedSigningAlgs, "RS256,ES256,PS256"))
	effective.ClockSkewSeconds = settings.OIDCConnectClockSkewSeconds
	if effective.ClockSkewSeconds == 0 {
		effective.ClockSkewSeconds = 120
	}
	effective.RequireEmailVerified = settings.OIDCConnectRequireEmailVerified
	effective.UserInfoEmailPath = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectUserInfoEmailPath, effective.UserInfoEmailPath))
	effective.UserInfoIDPath = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectUserInfoIDPath, effective.UserInfoIDPath))
	effective.UserInfoUsernamePath = strings.TrimSpace(firstNonEmpty(settings.OIDCConnectUserInfoUsernamePath, effective.UserInfoUsernamePath))

	if !effective.Enabled {
		return config.OIDCConnectConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "oauth login is disabled")
	}

	return effective, nil
}
