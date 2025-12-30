package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	RunModeStandard = "standard"
	RunModeSimple   = "simple"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Default      DefaultConfig      `mapstructure:"default"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Pricing      PricingConfig      `mapstructure:"pricing"`
	Gateway      GatewayConfig      `mapstructure:"gateway"`
	TokenRefresh TokenRefreshConfig `mapstructure:"token_refresh"`
	RunMode      string             `mapstructure:"run_mode" yaml:"run_mode"`
	Promotion    PromotionConfig    `mapstructure:"promotion"`
	Referral     ReferralConfig     `mapstructure:"referral"`
	Payment      PaymentConfig      `mapstructure:"payment"`
	Timezone     string             `mapstructure:"timezone"` // e.g. "Asia/Shanghai", "UTC"
	Gemini       GeminiConfig       `mapstructure:"gemini"`
}

type GeminiConfig struct {
	OAuth GeminiOAuthConfig `mapstructure:"oauth"`
}

type GeminiOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Scopes       string `mapstructure:"scopes"`
}

// TokenRefreshConfig OAuth token自动刷新配置
type TokenRefreshConfig struct {
	// 是否启用自动刷新
	Enabled bool `mapstructure:"enabled"`
	// 检查间隔（分钟）
	CheckIntervalMinutes int `mapstructure:"check_interval_minutes"`
	// 提前刷新时间（小时），在token过期前多久开始刷新
	RefreshBeforeExpiryHours float64 `mapstructure:"refresh_before_expiry_hours"`
	// 最大重试次数
	MaxRetries int `mapstructure:"max_retries"`
	// 重试退避基础时间（秒）
	RetryBackoffSeconds int `mapstructure:"retry_backoff_seconds"`
}

type PricingConfig struct {
	// 价格数据远程URL（默认使用LiteLLM镜像）
	RemoteURL string `mapstructure:"remote_url"`
	// 哈希校验文件URL
	HashURL string `mapstructure:"hash_url"`
	// 本地数据目录
	DataDir string `mapstructure:"data_dir"`
	// 回退文件路径
	FallbackFile string `mapstructure:"fallback_file"`
	// 更新间隔（小时）
	UpdateIntervalHours int `mapstructure:"update_interval_hours"`
	// 哈希校验间隔（分钟）
	HashCheckIntervalMinutes int `mapstructure:"hash_check_interval_minutes"`
}

type ServerConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	Mode              string `mapstructure:"mode"`                // debug/release
	ReadHeaderTimeout int    `mapstructure:"read_header_timeout"` // 读取请求头超时（秒）
	IdleTimeout       int    `mapstructure:"idle_timeout"`        // 空闲连接超时（秒）
}

// GatewayConfig API网关相关配置
type GatewayConfig struct {
	// 等待上游响应头的超时时间（秒），0表示无超时
	// 注意：这不影响流式数据传输，只控制等待响应头的时间
	ResponseHeaderTimeout int `mapstructure:"response_header_timeout"`
}

func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// DSNWithTimezone returns DSN with timezone setting
func (d *DatabaseConfig) DSNWithTimezone(tz string) string {
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode, tz,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

type DefaultConfig struct {
	AdminEmail      string  `mapstructure:"admin_email"`
	AdminPassword   string  `mapstructure:"admin_password"`
	UserConcurrency int     `mapstructure:"user_concurrency"`
	UserBalance     float64 `mapstructure:"user_balance"`
	ApiKeyPrefix    string  `mapstructure:"api_key_prefix"`
	RateMultiplier  float64 `mapstructure:"rate_multiplier"`
}

type PromotionConfig struct {
	Enabled       bool            `mapstructure:"enabled"`
	DurationHours int             `mapstructure:"duration_hours"`
	MinAmount     float64         `mapstructure:"min_amount"`
	Tiers         []PromotionTier `mapstructure:"tiers"`
}

type PromotionTier struct {
	Hours        int     `mapstructure:"hours"`
	BonusPercent float64 `mapstructure:"bonus_percent"`
}

type ReferralConfig struct {
	Enabled              bool    `mapstructure:"enabled"`
	LinkBaseURL          string  `mapstructure:"link_base_url"`
	RewardUSD            float64 `mapstructure:"reward_usd"`
	QualifiedRechargeCNY float64 `mapstructure:"qualified_recharge_cny"`
	QualifiedRechargeUSD float64 `mapstructure:"qualified_recharge_usd"`
	CodeLength           int     `mapstructure:"code_length"`
	MaxInviteesPerUser   int     `mapstructure:"max_invitees_per_user"`
}

type PaymentConfig struct {
	Enabled            bool             `mapstructure:"enabled"`
	BaseURL            string           `mapstructure:"base_url"`
	MinAmount          float64          `mapstructure:"min_amount"`
	MaxAmount          float64          `mapstructure:"max_amount"`
	ExchangeRate       float64          `mapstructure:"exchange_rate"`
	DiscountRate       float64          `mapstructure:"discount_rate"`
	OrderExpireMinutes int              `mapstructure:"order_expire_minutes"`
	MaxOrdersPerMinute int              `mapstructure:"max_orders_per_minute"`
	Packages           []PaymentPackage `mapstructure:"packages"`
	Zpay               ZpayConfig       `mapstructure:"zpay"`
	Stripe             StripeConfig     `mapstructure:"stripe"`
}

type PaymentPackage struct {
	AmountCNY float64 `mapstructure:"amount_cny"`
	AmountUSD float64 `mapstructure:"amount_usd"`
	Label     string  `mapstructure:"label"`
	Popular   bool    `mapstructure:"popular"`
}

type ZpayConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	PID             string `mapstructure:"pid"`
	Key             string `mapstructure:"key"`
	APIURL          string `mapstructure:"api_url"`
	SubmitURL       string `mapstructure:"submit_url"`
	QueryURL        string `mapstructure:"query_url"`
	PaymentMethods  string `mapstructure:"payment_methods"`
	OrderPrefix     string `mapstructure:"order_prefix"`
	NotifyURL       string `mapstructure:"notify_url"`
	ReturnURL       string `mapstructure:"return_url"`
	NotifyUser      bool   `mapstructure:"notify_user"`
	IPWhitelist     string `mapstructure:"ip_whitelist"`
	RequireHTTPS    bool   `mapstructure:"require_https"`
}

type StripeConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	APIKey        string `mapstructure:"api_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	APIVersion    string `mapstructure:"api_version"`
	PaymentMethods string `mapstructure:"payment_methods"`
	Currency      string `mapstructure:"currency"`
	SuccessURL    string `mapstructure:"success_url"`
	CancelURL     string `mapstructure:"cancel_url"`
	WechatClient  string `mapstructure:"wechat_client"`
	WechatAppID   string `mapstructure:"wechat_app_id"`
}

type RateLimitConfig struct {
	OverloadCooldownMinutes int `mapstructure:"overload_cooldown_minutes"` // 529过载冷却时间(分钟)
}

func NormalizeRunMode(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case RunModeStandard, RunModeSimple:
		return normalized
	default:
		return RunModeStandard
	}
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/sub2api")

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindLegacyEnvAliases(viper.GetViper())

	// 默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config error: %w", err)
		}
		// 配置文件不存在时使用默认值
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	cfg.RunMode = NormalizeRunMode(cfg.RunMode)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	return &cfg, nil
}

// bindLegacyEnvAliases binds legacy (non-namespaced) env vars to the current config keys.
//
// This repo primarily uses Viper's `.` -> `_` mapping, so config keys like:
//   payment.zpay.enabled -> PAYMENT_ZPAY_ENABLED
// But some deployments already use:
//   ZPAY_ENABLED / STRIPE_SECRET_KEY / ...
// This helper keeps backwards compatibility.
func bindLegacyEnvAliases(v *viper.Viper) {
	if v == nil {
		return
	}

	// ZPay
	_ = v.BindEnv("payment.zpay.enabled", "ZPAY_ENABLED")
	_ = v.BindEnv("payment.zpay.pid", "ZPAY_PID")
	_ = v.BindEnv("payment.zpay.key", "ZPAY_KEY")
	_ = v.BindEnv("payment.zpay.api_url", "ZPAY_API_URL")
	_ = v.BindEnv("payment.zpay.submit_url", "ZPAY_SUBMIT_URL")
	_ = v.BindEnv("payment.zpay.query_url", "ZPAY_QUERY_URL")
	_ = v.BindEnv("payment.zpay.payment_methods", "ZPAY_PAYMENT_METHODS")
	_ = v.BindEnv("payment.zpay.order_prefix", "ZPAY_ORDER_PREFIX")
	_ = v.BindEnv("payment.zpay.notify_url", "ZPAY_NOTIFY_URL")
	_ = v.BindEnv("payment.zpay.return_url", "ZPAY_RETURN_URL")
	_ = v.BindEnv("payment.zpay.notify_user", "ZPAY_NOTIFY_USER")
	_ = v.BindEnv("payment.zpay.ip_whitelist", "ZPAY_IP_WHITELIST")
	_ = v.BindEnv("payment.zpay.require_https", "ZPAY_REQUIRE_HTTPS")

	// Stripe
	_ = v.BindEnv("payment.stripe.enabled", "STRIPE_ENABLED")
	// Common naming: STRIPE_SECRET_KEY is the Stripe API key.
	_ = v.BindEnv("payment.stripe.api_key", "STRIPE_API_KEY", "STRIPE_SECRET_KEY")
	_ = v.BindEnv("payment.stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET")
	_ = v.BindEnv("payment.stripe.api_version", "STRIPE_API_VERSION")
	_ = v.BindEnv("payment.stripe.payment_methods", "STRIPE_PAYMENT_METHODS")
	_ = v.BindEnv("payment.stripe.currency", "STRIPE_CURRENCY")
	_ = v.BindEnv("payment.stripe.success_url", "STRIPE_SUCCESS_URL")
	_ = v.BindEnv("payment.stripe.cancel_url", "STRIPE_CANCEL_URL")
	_ = v.BindEnv("payment.stripe.wechat_client", "STRIPE_WECHAT_CLIENT")
	_ = v.BindEnv("payment.stripe.wechat_app_id", "STRIPE_WECHAT_APP_ID")
}

func setDefaults() {
	viper.SetDefault("run_mode", RunModeStandard)

	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_header_timeout", 30) // 30秒读取请求头
	viper.SetDefault("server.idle_timeout", 120)       // 120秒空闲超时

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "sub2api")
	viper.SetDefault("database.sslmode", "disable")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.expire_hour", 24)

	// Default
	viper.SetDefault("default.admin_email", "admin@sub2api.com")
	viper.SetDefault("default.admin_password", "admin123")
	viper.SetDefault("default.user_concurrency", 5)
	viper.SetDefault("default.user_balance", 0)
	viper.SetDefault("default.api_key_prefix", "sk-")
	viper.SetDefault("default.rate_multiplier", 1.0)

	// Promotion
	viper.SetDefault("promotion.enabled", false)
	viper.SetDefault("promotion.duration_hours", 72)
	viper.SetDefault("promotion.min_amount", 0.0)
	viper.SetDefault("promotion.tiers", []map[string]any{
		{"hours": 24, "bonus_percent": 30},
		{"hours": 36, "bonus_percent": 20},
		{"hours": 48, "bonus_percent": 10},
		{"hours": 72, "bonus_percent": 5},
	})

	// Referral
	viper.SetDefault("referral.enabled", false)
	viper.SetDefault("referral.link_base_url", "")
	viper.SetDefault("referral.reward_usd", 0.0)
	viper.SetDefault("referral.qualified_recharge_cny", 0.0)
	viper.SetDefault("referral.qualified_recharge_usd", 0.0)
	viper.SetDefault("referral.code_length", 8)
	viper.SetDefault("referral.max_invitees_per_user", 0)

	// Payment
	viper.SetDefault("payment.enabled", false)
	viper.SetDefault("payment.base_url", "")
	viper.SetDefault("payment.min_amount", 1.0)
	viper.SetDefault("payment.max_amount", 10000.0)
	viper.SetDefault("payment.exchange_rate", 7.2)
	viper.SetDefault("payment.discount_rate", 1.0)
	viper.SetDefault("payment.order_expire_minutes", 30)
	viper.SetDefault("payment.max_orders_per_minute", 3)
	viper.SetDefault("payment.packages", []map[string]any{
		{"amount_usd": 100, "label": "$100", "popular": false},
		{"amount_usd": 200, "label": "$200", "popular": false},
		{"amount_usd": 500, "label": "$500", "popular": true},
		{"amount_usd": 1000, "label": "$1000", "popular": false},
	})
	viper.SetDefault("payment.zpay.enabled", false)
	viper.SetDefault("payment.zpay.pid", "")
	viper.SetDefault("payment.zpay.key", "")
	viper.SetDefault("payment.zpay.api_url", "https://zpayz.cn")
	viper.SetDefault("payment.zpay.submit_url", "https://zpayz.cn/submit.php")
	viper.SetDefault("payment.zpay.query_url", "https://zpayz.cn/api.php")
	viper.SetDefault("payment.zpay.payment_methods", "alipay,wxpay")
	viper.SetDefault("payment.zpay.order_prefix", "")
	// Notify can be either /api/payment/webhook/zpay or /payment/zpay/notify (both are routed).
	viper.SetDefault("payment.zpay.notify_url", "/payment/zpay/notify")
	viper.SetDefault("payment.zpay.return_url", "/payment/zpay/return")
	viper.SetDefault("payment.zpay.notify_user", false)
	viper.SetDefault("payment.zpay.ip_whitelist", "")
	viper.SetDefault("payment.zpay.require_https", true)

	viper.SetDefault("payment.stripe.enabled", false)
	viper.SetDefault("payment.stripe.api_key", "")
	viper.SetDefault("payment.stripe.webhook_secret", "")
	viper.SetDefault("payment.stripe.api_version", "")
	viper.SetDefault("payment.stripe.payment_methods", "wechat_pay")
	viper.SetDefault("payment.stripe.currency", "CNY")
	// Currently these URLs are mainly for compatibility with existing deployments.
	viper.SetDefault("payment.stripe.success_url", "/payment/return/stripe?order={ORDER_ID}&status=success")
	viper.SetDefault("payment.stripe.cancel_url", "/payment/return/stripe?order={ORDER_ID}&status=cancel")
	viper.SetDefault("payment.stripe.wechat_client", "web")
	viper.SetDefault("payment.stripe.wechat_app_id", "")

	// RateLimit
	viper.SetDefault("rate_limit.overload_cooldown_minutes", 10)

	// Pricing - 从 price-mirror 分支同步，该分支维护了 sha256 哈希文件用于增量更新检查
	viper.SetDefault("pricing.remote_url", "https://raw.githubusercontent.com/Wei-Shaw/claude-relay-service/price-mirror/model_prices_and_context_window.json")
	viper.SetDefault("pricing.hash_url", "https://raw.githubusercontent.com/Wei-Shaw/claude-relay-service/price-mirror/model_prices_and_context_window.sha256")
	viper.SetDefault("pricing.data_dir", "./data")
	viper.SetDefault("pricing.fallback_file", "./resources/model-pricing/model_prices_and_context_window.json")
	viper.SetDefault("pricing.update_interval_hours", 24)
	viper.SetDefault("pricing.hash_check_interval_minutes", 10)

	// Timezone (default to Asia/Shanghai for Chinese users)
	viper.SetDefault("timezone", "Asia/Shanghai")

	// Gateway
	viper.SetDefault("gateway.response_header_timeout", 300) // 300秒(5分钟)等待上游响应头，LLM高负载时可能排队较久

	// TokenRefresh
	viper.SetDefault("token_refresh.enabled", true)
	viper.SetDefault("token_refresh.check_interval_minutes", 5)        // 每5分钟检查一次
	viper.SetDefault("token_refresh.refresh_before_expiry_hours", 0.5) // 提前30分钟刷新（适配Google 1小时token）
	viper.SetDefault("token_refresh.max_retries", 3)                   // 最多重试3次
	viper.SetDefault("token_refresh.retry_backoff_seconds", 2)         // 重试退避基础2秒

	// Gemini OAuth - configure via environment variables or config file
	// GEMINI_OAUTH_CLIENT_ID and GEMINI_OAUTH_CLIENT_SECRET
	// Default: uses Gemini CLI public credentials (set via environment)
	viper.SetDefault("gemini.oauth.client_id", "")
	viper.SetDefault("gemini.oauth.client_secret", "")
	viper.SetDefault("gemini.oauth.scopes", "")
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if c.JWT.Secret == "change-me-in-production" && c.Server.Mode == "release" {
		return fmt.Errorf("jwt.secret must be changed in production")
	}
	return nil
}

// GetServerAddress returns the server address (host:port) from config file or environment variable.
// This is a lightweight function that can be used before full config validation,
// such as during setup wizard startup.
// Priority: config.yaml > environment variables > defaults
func GetServerAddress() string {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/sub2api")

	// Support SERVER_HOST and SERVER_PORT environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// Try to read config file (ignore errors if not found)
	_ = v.ReadInConfig()

	host := v.GetString("server.host")
	port := v.GetInt("server.port")
	return fmt.Sprintf("%s:%d", host, port)
}
