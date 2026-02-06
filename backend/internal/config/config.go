// Package config provides configuration loading, defaults, and validation.
package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	RunModeStandard = "standard"
	RunModeSimple   = "simple"
)

// DefaultCSPPolicy is the default Content-Security-Policy with nonce support
// __CSP_NONCE__ will be replaced with actual nonce at request time by the SecurityHeaders middleware
const DefaultCSPPolicy = "default-src 'self'; script-src 'self' __CSP_NONCE__ https://challenges.cloudflare.com https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' https:; frame-src https://challenges.cloudflare.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

// 连接池隔离策略常量
// 用于控制上游 HTTP 连接池的隔离粒度，影响连接复用和资源消耗
const (
	// ConnectionPoolIsolationProxy: 按代理隔离
	// 同一代理地址共享连接池，适合代理数量少、账户数量多的场景
	ConnectionPoolIsolationProxy = "proxy"
	// ConnectionPoolIsolationAccount: 按账户隔离
	// 每个账户独立连接池，适合账户数量少、需要严格隔离的场景
	ConnectionPoolIsolationAccount = "account"
	// ConnectionPoolIsolationAccountProxy: 按账户+代理组合隔离（默认）
	// 同一账户+代理组合共享连接池，提供最细粒度的隔离
	ConnectionPoolIsolationAccountProxy = "account_proxy"
)

type Config struct {
	Server       ServerConfig               `mapstructure:"server"`
	CORS         CORSConfig                 `mapstructure:"cors"`
	Security     SecurityConfig             `mapstructure:"security"`
	Billing      BillingConfig              `mapstructure:"billing"`
	Turnstile    TurnstileConfig            `mapstructure:"turnstile"`
	Database     DatabaseConfig             `mapstructure:"database"`
	Redis        RedisConfig                `mapstructure:"redis"`
	Ops          OpsConfig                  `mapstructure:"ops"`
	APIKeyAuth   APIKeyAuthCacheConfig      `mapstructure:"api_key_auth_cache"`
	Dashboard    DashboardCacheConfig       `mapstructure:"dashboard_cache"`
	DashboardAgg DashboardAggregationConfig `mapstructure:"dashboard_aggregation"`
	JWT          JWTConfig                  `mapstructure:"jwt"`
	LinuxDo      LinuxDoConnectConfig       `mapstructure:"linuxdo_connect"`
	Default      DefaultConfig              `mapstructure:"default"`
	RateLimit    RateLimitConfig            `mapstructure:"rate_limit"`
	Pricing      PricingConfig              `mapstructure:"pricing"`
	Gateway      GatewayConfig              `mapstructure:"gateway"`
	UsageCleanup UsageCleanupConfig         `mapstructure:"usage_cleanup"`
	Concurrency  ConcurrencyConfig          `mapstructure:"concurrency"`
	TokenRefresh TokenRefreshConfig         `mapstructure:"token_refresh"`
	RunMode      string                     `mapstructure:"run_mode" yaml:"run_mode"`
	Promotion    PromotionConfig            `mapstructure:"promotion"`
	Referral     ReferralConfig             `mapstructure:"referral"`
	Payment      PaymentConfig              `mapstructure:"payment"`
	Dingtalk     DingtalkConfig             `mapstructure:"dingtalk"`
	DingtalkBot  DingtalkBotConfig          `mapstructure:"dingtalk_bot"`
	Timezone     string                     `mapstructure:"timezone"` // e.g. "Asia/Shanghai", "UTC"
	Gemini       GeminiConfig               `mapstructure:"gemini"`
	Update       UpdateConfig               `mapstructure:"update"`
}

type GeminiConfig struct {
	OAuth GeminiOAuthConfig `mapstructure:"oauth"`
	Quota GeminiQuotaConfig `mapstructure:"quota"`
}

type GeminiOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Scopes       string `mapstructure:"scopes"`
}

type GeminiQuotaConfig struct {
	Tiers  map[string]GeminiTierQuotaConfig `mapstructure:"tiers"`
	Policy string                           `mapstructure:"policy"`
}

type GeminiTierQuotaConfig struct {
	ProRPD          *int64 `mapstructure:"pro_rpd" json:"pro_rpd"`
	FlashRPD        *int64 `mapstructure:"flash_rpd" json:"flash_rpd"`
	CooldownMinutes *int   `mapstructure:"cooldown_minutes" json:"cooldown_minutes"`
}

type UpdateConfig struct {
	// ProxyURL 用于访问 GitHub 的代理地址
	// 支持 http/https/socks5/socks5h 协议
	// 例如: "http://127.0.0.1:7890", "socks5://127.0.0.1:1080"
	ProxyURL string `mapstructure:"proxy_url"`
}

type LinuxDoConnectConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	ClientID            string `mapstructure:"client_id"`
	ClientSecret        string `mapstructure:"client_secret"`
	AuthorizeURL        string `mapstructure:"authorize_url"`
	TokenURL            string `mapstructure:"token_url"`
	UserInfoURL         string `mapstructure:"userinfo_url"`
	Scopes              string `mapstructure:"scopes"`
	RedirectURL         string `mapstructure:"redirect_url"`          // 后端回调地址（需在提供方后台登记）
	FrontendRedirectURL string `mapstructure:"frontend_redirect_url"` // 前端接收 token 的路由（默认：/auth/linuxdo/callback）
	TokenAuthMethod     string `mapstructure:"token_auth_method"`     // client_secret_post / client_secret_basic / none
	UsePKCE             bool   `mapstructure:"use_pkce"`

	// 可选：用于从 userinfo JSON 中提取字段的 gjson 路径。
	// 为空时，服务端会尝试一组常见字段名。
	UserInfoEmailPath    string `mapstructure:"userinfo_email_path"`
	UserInfoIDPath       string `mapstructure:"userinfo_id_path"`
	UserInfoUsernamePath string `mapstructure:"userinfo_username_path"`
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
	Host              string   `mapstructure:"host"`
	Port              int      `mapstructure:"port"`
	Mode              string   `mapstructure:"mode"`                // debug/release
	ReadHeaderTimeout int      `mapstructure:"read_header_timeout"` // 读取请求头超时（秒）
	IdleTimeout       int      `mapstructure:"idle_timeout"`        // 空闲连接超时（秒）
	TrustedProxies    []string `mapstructure:"trusted_proxies"`     // 可信代理列表（CIDR/IP）
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type SecurityConfig struct {
	URLAllowlist    URLAllowlistConfig   `mapstructure:"url_allowlist"`
	ResponseHeaders ResponseHeaderConfig `mapstructure:"response_headers"`
	CSP             CSPConfig            `mapstructure:"csp"`
	ProxyProbe      ProxyProbeConfig     `mapstructure:"proxy_probe"`
	// AdminAPIKeyReadOnly controls the allowlist policy for the read-only Admin API Key (x-api-key: admin-ro-...).
	AdminAPIKeyReadOnly AdminAPIKeyReadOnlyConfig `mapstructure:"admin_api_key_read_only"`
}

type URLAllowlistConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	UpstreamHosts     []string `mapstructure:"upstream_hosts"`
	PricingHosts      []string `mapstructure:"pricing_hosts"`
	CRSHosts          []string `mapstructure:"crs_hosts"`
	AllowPrivateHosts bool     `mapstructure:"allow_private_hosts"`
	// 关闭 URL 白名单校验时，是否允许 http URL（默认只允许 https）
	AllowInsecureHTTP bool `mapstructure:"allow_insecure_http"`
}

type ResponseHeaderConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	AdditionalAllowed []string `mapstructure:"additional_allowed"`
	ForceRemove       []string `mapstructure:"force_remove"`
}

type CSPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Policy  string `mapstructure:"policy"`
}

type ProxyProbeConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"` // 已禁用：禁止跳过 TLS 证书验证
}

type AdminAPIKeyReadOnlyConfig struct {
	// AllowedPaths defines the exact request paths that the read-only Admin API key can access.
	// Only GET requests are allowed by the middleware.
	AllowedPaths []string `mapstructure:"allowed_paths"`
	// AllowedPathPrefixes defines path prefixes (with segment boundary matching) that the read-only Admin API key can access.
	// For example: "/api/v1/admin/users" matches "/api/v1/admin/users" and "/api/v1/admin/users/123", but not "/api/v1/admin/users2".
	AllowedPathPrefixes []string `mapstructure:"allowed_path_prefixes"`
}

type BillingConfig struct {
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	Enabled             bool `mapstructure:"enabled"`
	FailureThreshold    int  `mapstructure:"failure_threshold"`
	ResetTimeoutSeconds int  `mapstructure:"reset_timeout_seconds"`
	HalfOpenRequests    int  `mapstructure:"half_open_requests"`
}

type ConcurrencyConfig struct {
	// PingInterval: 并发等待期间的 SSE ping 间隔（秒）
	PingInterval int `mapstructure:"ping_interval"`
}

// GatewayConfig API网关相关配置
type GatewayConfig struct {
	// 等待上游响应头的超时时间（秒），0表示无超时
	// 注意：这不影响流式数据传输，只控制等待响应头的时间
	ResponseHeaderTimeout int `mapstructure:"response_header_timeout"`
	// 请求体最大字节数，用于网关请求体大小限制
	MaxBodySize int64 `mapstructure:"max_body_size"`
	// ConnectionPoolIsolation: 上游连接池隔离策略（proxy/account/account_proxy）
	ConnectionPoolIsolation string `mapstructure:"connection_pool_isolation"`

	// HTTP 上游连接池配置（性能优化：支持高并发场景调优）
	// MaxIdleConns: 所有主机的最大空闲连接总数
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// MaxIdleConnsPerHost: 每个主机的最大空闲连接数（关键参数，影响连接复用率）
	MaxIdleConnsPerHost int `mapstructure:"max_idle_conns_per_host"`
	// MaxConnsPerHost: 每个主机的最大连接数（包括活跃+空闲），0表示无限制
	MaxConnsPerHost int `mapstructure:"max_conns_per_host"`
	// IdleConnTimeoutSeconds: 空闲连接超时时间（秒）
	IdleConnTimeoutSeconds int `mapstructure:"idle_conn_timeout_seconds"`
	// MaxUpstreamClients: 上游连接池客户端最大缓存数量
	// 当使用连接池隔离策略时，系统会为不同的账户/代理组合创建独立的 HTTP 客户端
	// 此参数限制缓存的客户端数量，超出后会淘汰最久未使用的客户端
	// 建议值：预估的活跃账户数 * 1.2（留有余量）
	MaxUpstreamClients int `mapstructure:"max_upstream_clients"`
	// ClientIdleTTLSeconds: 上游连接池客户端空闲回收阈值（秒）
	// 超过此时间未使用的客户端会被标记为可回收
	// 建议值：根据用户访问频率设置，一般 10-30 分钟
	ClientIdleTTLSeconds int `mapstructure:"client_idle_ttl_seconds"`
	// ConcurrencySlotTTLMinutes: 并发槽位过期时间（分钟）
	// 应大于最长 LLM 请求时间，防止请求完成前槽位过期
	ConcurrencySlotTTLMinutes int `mapstructure:"concurrency_slot_ttl_minutes"`
	// SessionIdleTimeoutMinutes: 会话空闲超时时间（分钟），默认 5 分钟
	// 用于 Anthropic OAuth/SetupToken 账号的会话数量限制功能
	// 空闲超过此时间的会话将被自动释放
	SessionIdleTimeoutMinutes int `mapstructure:"session_idle_timeout_minutes"`

	// StreamDataIntervalTimeout: 流数据间隔超时（秒），0表示禁用
	StreamDataIntervalTimeout int `mapstructure:"stream_data_interval_timeout"`
	// StreamKeepaliveInterval: 流式 keepalive 间隔（秒），0表示禁用
	StreamKeepaliveInterval int `mapstructure:"stream_keepalive_interval"`
	// MaxLineSize: 上游 SSE 单行最大字节数（0使用默认值）
	MaxLineSize int `mapstructure:"max_line_size"`

	// 是否记录上游错误响应体摘要（避免输出请求内容）
	LogUpstreamErrorBody bool `mapstructure:"log_upstream_error_body"`
	// 上游错误响应体记录最大字节数（超过会截断）
	LogUpstreamErrorBodyMaxBytes int `mapstructure:"log_upstream_error_body_max_bytes"`

	// API-key 账号在客户端未提供 anthropic-beta 时，是否按需自动补齐（默认关闭以保持兼容）
	InjectBetaForAPIKey bool `mapstructure:"inject_beta_for_apikey"`

	// 是否允许对部分 400 错误触发 failover（默认关闭以避免改变语义）
	FailoverOn400 bool `mapstructure:"failover_on_400"`

	// 账户切换最大次数（遇到上游错误时切换到其他账户的次数上限）
	MaxAccountSwitches int `mapstructure:"max_account_switches"`
	// Gemini 账户切换最大次数（Gemini 平台单独配置，因 API 限制更严格）
	MaxAccountSwitchesGemini int `mapstructure:"max_account_switches_gemini"`

	// Antigravity 429 fallback 限流时间（分钟），解析重置时间失败时使用
	AntigravityFallbackCooldownMinutes int `mapstructure:"antigravity_fallback_cooldown_minutes"`

	// Scheduling: 账号调度相关配置
	Scheduling GatewaySchedulingConfig `mapstructure:"scheduling"`

	// TLSFingerprint: TLS指纹伪装配置
	TLSFingerprint TLSFingerprintConfig `mapstructure:"tls_fingerprint"`
}

// TLSFingerprintConfig TLS指纹伪装配置
// 用于模拟 Claude CLI (Node.js) 的 TLS 握手特征，避免被识别为非官方客户端
type TLSFingerprintConfig struct {
	// Enabled: 是否全局启用TLS指纹功能
	Enabled bool `mapstructure:"enabled"`
	// Profiles: 预定义的TLS指纹配置模板
	// key 为模板名称，如 "claude_cli_v2", "chrome_120" 等
	Profiles map[string]TLSProfileConfig `mapstructure:"profiles"`
}

// TLSProfileConfig 单个TLS指纹模板的配置
type TLSProfileConfig struct {
	// Name: 模板显示名称
	Name string `mapstructure:"name"`
	// EnableGREASE: 是否启用GREASE扩展（Chrome使用，Node.js不使用）
	EnableGREASE bool `mapstructure:"enable_grease"`
	// CipherSuites: TLS加密套件列表（空则使用内置默认值）
	CipherSuites []uint16 `mapstructure:"cipher_suites"`
	// Curves: 椭圆曲线列表（空则使用内置默认值）
	Curves []uint16 `mapstructure:"curves"`
	// PointFormats: 点格式列表（空则使用内置默认值）
	PointFormats []uint8 `mapstructure:"point_formats"`
}

// GatewaySchedulingConfig accounts scheduling configuration.
type GatewaySchedulingConfig struct {
	// 粘性会话排队配置
	StickySessionMaxWaiting  int           `mapstructure:"sticky_session_max_waiting"`
	StickySessionWaitTimeout time.Duration `mapstructure:"sticky_session_wait_timeout"`

	// 兜底排队配置
	FallbackWaitTimeout time.Duration `mapstructure:"fallback_wait_timeout"`
	FallbackMaxWaiting  int           `mapstructure:"fallback_max_waiting"`

	// 兜底层账户选择策略: "last_used"(按最后使用时间排序，默认) 或 "random"(随机)
	FallbackSelectionMode string `mapstructure:"fallback_selection_mode"`

	// 负载计算
	LoadBatchEnabled bool `mapstructure:"load_batch_enabled"`

	// 过期槽位清理周期（0 表示禁用）
	SlotCleanupInterval time.Duration `mapstructure:"slot_cleanup_interval"`

	// 受控回源配置
	DbFallbackEnabled bool `mapstructure:"db_fallback_enabled"`
	// 受控回源超时（秒），0 表示不额外收紧超时
	DbFallbackTimeoutSeconds int `mapstructure:"db_fallback_timeout_seconds"`
	// 受控回源限流（实例级 QPS），0 表示不限制
	DbFallbackMaxQPS int `mapstructure:"db_fallback_max_qps"`

	// Outbox 轮询与滞后阈值配置
	// Outbox 轮询周期（秒）
	OutboxPollIntervalSeconds int `mapstructure:"outbox_poll_interval_seconds"`
	// Outbox 滞后告警阈值（秒）
	OutboxLagWarnSeconds int `mapstructure:"outbox_lag_warn_seconds"`
	// Outbox 触发强制重建阈值（秒）
	OutboxLagRebuildSeconds int `mapstructure:"outbox_lag_rebuild_seconds"`
	// Outbox 连续滞后触发次数
	OutboxLagRebuildFailures int `mapstructure:"outbox_lag_rebuild_failures"`
	// Outbox 积压触发重建阈值（行数）
	OutboxBacklogRebuildRows int `mapstructure:"outbox_backlog_rebuild_rows"`

	// 全量重建周期配置
	// 全量重建周期（秒），0 表示禁用
	FullRebuildIntervalSeconds int `mapstructure:"full_rebuild_interval_seconds"`

	// AnthropicAPIKeyMonitor: Anthropic API-key 账号连通性监控（自动启停调度）
	AnthropicAPIKeyMonitor AnthropicAPIKeyMonitorConfig `mapstructure:"anthropic_apikey_monitor"`
}

// AnthropicAPIKeyMonitorConfig controls the background connectivity monitor for Anthropic API-key accounts.
//
// When enabled, the service periodically performs a lightweight "test account connection" call against
// the configured Anthropic upstream (account.credentials.base_url), and:
// - disables scheduling after N consecutive failures
// - re-enables scheduling after N consecutive successes
type AnthropicAPIKeyMonitorConfig struct {
	Enabled bool `mapstructure:"enabled"`

	// Interval between checks. Recommended: 10s.
	Interval time.Duration `mapstructure:"interval"`

	// FailureThreshold: consecutive failures required to stop scheduling. Recommended: 6.
	FailureThreshold int `mapstructure:"failure_threshold"`
	// SuccessThreshold: consecutive successes required to resume scheduling. Recommended: 6.
	SuccessThreshold int `mapstructure:"success_threshold"`

	// RequestTimeout bounds a single upstream test request. Recommended: 8s.
	RequestTimeout time.Duration `mapstructure:"request_timeout"`

	// MaxConcurrency limits concurrent upstream tests per cycle. 0 uses a safe default.
	MaxConcurrency int `mapstructure:"max_concurrency"`

	// ModelID optionally overrides the model used for the test request.
	// Empty uses the backend default (claude.DefaultMonitorModel), then applies account model_mapping.
	ModelID string `mapstructure:"model_id"`

	// IncludeAccountIDs optionally restricts the monitor to specific account IDs.
	// When non-empty, only matching Anthropic API-key active accounts will be tested.
	IncludeAccountIDs []int64 `mapstructure:"include_account_ids"`
	// ExcludeAccountIDs optionally skips specific account IDs from monitoring.
	ExcludeAccountIDs []int64 `mapstructure:"exclude_account_ids"`
}

func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig 数据库连接配置
// 性能优化：新增连接池参数，避免频繁创建/销毁连接
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	// 连接池配置（性能优化：可配置化连接池参数）
	// MaxOpenConns: 最大打开连接数，控制数据库连接上限，防止资源耗尽
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// MaxIdleConns: 最大空闲连接数，保持热连接减少建连延迟
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// ConnMaxLifetimeMinutes: 连接最大存活时间，防止长连接导致的资源泄漏
	ConnMaxLifetimeMinutes int `mapstructure:"conn_max_lifetime_minutes"`
	// ConnMaxIdleTimeMinutes: 空闲连接最大存活时间，及时释放不活跃连接
	ConnMaxIdleTimeMinutes int `mapstructure:"conn_max_idle_time_minutes"`
}

func (d *DatabaseConfig) DSN() string {
	// 当密码为空时不包含 password 参数，避免 libpq 解析错误
	if d.Password == "" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=%s",
			d.Host, d.Port, d.User, d.DBName, d.SSLMode,
		)
	}
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
	// 当密码为空时不包含 password 参数，避免 libpq 解析错误
	if d.Password == "" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=%s TimeZone=%s",
			d.Host, d.Port, d.User, d.DBName, d.SSLMode, tz,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode, tz,
	)
}

// RedisConfig Redis 连接配置
// 性能优化：新增连接池和超时参数，提升高并发场景下的吞吐量
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	// 连接池与超时配置（性能优化：可配置化连接池参数）
	// DialTimeoutSeconds: 建立连接超时，防止慢连接阻塞
	DialTimeoutSeconds int `mapstructure:"dial_timeout_seconds"`
	// ReadTimeoutSeconds: 读取超时，避免慢查询阻塞连接池
	ReadTimeoutSeconds int `mapstructure:"read_timeout_seconds"`
	// WriteTimeoutSeconds: 写入超时，避免慢写入阻塞连接池
	WriteTimeoutSeconds int `mapstructure:"write_timeout_seconds"`
	// PoolSize: 连接池大小，控制最大并发连接数
	PoolSize int `mapstructure:"pool_size"`
	// MinIdleConns: 最小空闲连接数，保持热连接减少冷启动延迟
	MinIdleConns int `mapstructure:"min_idle_conns"`
}

func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type OpsConfig struct {
	// Enabled controls whether ops features should run.
	//
	// NOTE: vNext still has a DB-backed feature flag (ops_monitoring_enabled) for runtime on/off.
	// This config flag is the "hard switch" for deployments that want to disable ops completely.
	Enabled bool `mapstructure:"enabled"`

	// UsePreaggregatedTables prefers ops_metrics_hourly/daily for long-window dashboard queries.
	UsePreaggregatedTables bool `mapstructure:"use_preaggregated_tables"`

	// Cleanup controls periodic deletion of old ops data to prevent unbounded growth.
	Cleanup OpsCleanupConfig `mapstructure:"cleanup"`

	// MetricsCollectorCache controls Redis caching for expensive per-window collector queries.
	MetricsCollectorCache OpsMetricsCollectorCacheConfig `mapstructure:"metrics_collector_cache"`

	// Pre-aggregation configuration.
	Aggregation OpsAggregationConfig `mapstructure:"aggregation"`
}

type OpsCleanupConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Schedule string `mapstructure:"schedule"`

	// Retention days (0 disables that cleanup target).
	//
	// vNext requirement: default 30 days across ops datasets.
	ErrorLogRetentionDays      int `mapstructure:"error_log_retention_days"`
	MinuteMetricsRetentionDays int `mapstructure:"minute_metrics_retention_days"`
	HourlyMetricsRetentionDays int `mapstructure:"hourly_metrics_retention_days"`
}

type OpsAggregationConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type OpsMetricsCollectorCacheConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	TTL     time.Duration `mapstructure:"ttl"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

type TurnstileConfig struct {
	Required bool `mapstructure:"required"`
}

type DefaultConfig struct {
	AdminEmail      string  `mapstructure:"admin_email"`
	AdminPassword   string  `mapstructure:"admin_password"`
	UserConcurrency int     `mapstructure:"user_concurrency"`
	UserBalance     float64 `mapstructure:"user_balance"`
	APIKeyPrefix    string  `mapstructure:"api_key_prefix"`
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
	Enabled                bool     `mapstructure:"enabled"`
	BaseURL                string   `mapstructure:"base_url"`
	LinkBaseURL            string   `mapstructure:"link_base_url"`
	RewardUSD              float64  `mapstructure:"reward_usd"`
	QualifiedRechargeCNY   float64  `mapstructure:"qualified_recharge_cny"`
	QualifiedRechargeUSD   float64  `mapstructure:"qualified_recharge_usd"`
	QualifiedRechargeTypes []string `mapstructure:"qualified_recharge_types"`
	CodeLength             int      `mapstructure:"code_length"`
	MaxInviteesPerUser     int      `mapstructure:"max_invitees_per_user"`
}

type DingtalkConfig struct {
	Enabled              bool   `mapstructure:"enabled"`
	Env                  string `mapstructure:"env"`
	WebhookURL           string `mapstructure:"webhook_url"`
	Secret               string `mapstructure:"secret"`
	AtMobiles            string `mapstructure:"at_mobiles"` // comma separated
	AtAll                bool   `mapstructure:"at_all"`
	PaymentNotifyEnabled bool   `mapstructure:"payment_notify_enabled"`
}

type DingtalkBotConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	AccessToken      string `mapstructure:"access_token"`
	SignSecret       string `mapstructure:"sign_secret"`
	AllowedSenderIDs string `mapstructure:"allowed_sender_ids"` // comma separated
	DefaultRemark    string `mapstructure:"default_remark"`
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
	OrderPrefix        string           `mapstructure:"order_prefix"`
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
	Enabled        bool   `mapstructure:"enabled"`
	PID            string `mapstructure:"pid"`
	Key            string `mapstructure:"key"`
	APIURL         string `mapstructure:"api_url"`
	SubmitURL      string `mapstructure:"submit_url"`
	QueryURL       string `mapstructure:"query_url"`
	PaymentMethods string `mapstructure:"payment_methods"`
	OrderPrefix    string `mapstructure:"order_prefix"`
	NotifyURL      string `mapstructure:"notify_url"`
	ReturnURL      string `mapstructure:"return_url"`
	NotifyUser     bool   `mapstructure:"notify_user"`
	IPWhitelist    string `mapstructure:"ip_whitelist"`
	RequireHTTPS   bool   `mapstructure:"require_https"`

	// Multi-channel support: specify channel ID (cid) for different payment methods
	// When you have multiple payment channels in ZPAY console (e.g., multiple WeChat channels),
	// use these to specify which channel to use for each payment type
	AlipayChannelID string `mapstructure:"alipay_channel_id"` // Alipay channel ID (optional)
	WechatChannelID string `mapstructure:"wechat_channel_id"` // WeChat channel ID (optional)
}

type StripeConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	APIKey         string `mapstructure:"api_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	APIVersion     string `mapstructure:"api_version"`
	PaymentMethods string `mapstructure:"payment_methods"`
	Currency       string `mapstructure:"currency"`
	SuccessURL     string `mapstructure:"success_url"`
	CancelURL      string `mapstructure:"cancel_url"`
	WechatClient   string `mapstructure:"wechat_client"`
	WechatAppID    string `mapstructure:"wechat_app_id"`
}

type RateLimitConfig struct {
	OverloadCooldownMinutes int `mapstructure:"overload_cooldown_minutes"` // 529过载冷却时间(分钟)
	// FallbackCooldownMinutes is the default cooldown (minutes) applied when an upstream 429 response
	// does not provide a usable reset time (e.g. missing/invalid headers).
	//
	// <= 0 means use the built-in default (5 minutes).
	FallbackCooldownMinutes int `mapstructure:"fallback_cooldown_minutes"` // 429兜底冷却时间(分钟)
}

// APIKeyAuthCacheConfig API Key 认证缓存配置
type APIKeyAuthCacheConfig struct {
	L1Size             int  `mapstructure:"l1_size"`
	L1TTLSeconds       int  `mapstructure:"l1_ttl_seconds"`
	L2TTLSeconds       int  `mapstructure:"l2_ttl_seconds"`
	NegativeTTLSeconds int  `mapstructure:"negative_ttl_seconds"`
	JitterPercent      int  `mapstructure:"jitter_percent"`
	Singleflight       bool `mapstructure:"singleflight"`
}

// DashboardCacheConfig 仪表盘统计缓存配置
type DashboardCacheConfig struct {
	// Enabled: 是否启用仪表盘缓存
	Enabled bool `mapstructure:"enabled"`
	// KeyPrefix: Redis key 前缀，用于多环境隔离
	KeyPrefix string `mapstructure:"key_prefix"`
	// StatsFreshTTLSeconds: 缓存命中认为“新鲜”的时间窗口（秒）
	StatsFreshTTLSeconds int `mapstructure:"stats_fresh_ttl_seconds"`
	// StatsTTLSeconds: Redis 缓存总 TTL（秒）
	StatsTTLSeconds int `mapstructure:"stats_ttl_seconds"`
	// StatsRefreshTimeoutSeconds: 异步刷新超时（秒）
	StatsRefreshTimeoutSeconds int `mapstructure:"stats_refresh_timeout_seconds"`
}

// DashboardAggregationConfig 仪表盘预聚合配置
type DashboardAggregationConfig struct {
	// Enabled: 是否启用预聚合作业
	Enabled bool `mapstructure:"enabled"`
	// IntervalSeconds: 聚合刷新间隔（秒）
	IntervalSeconds int `mapstructure:"interval_seconds"`
	// LookbackSeconds: 回看窗口（秒）
	LookbackSeconds int `mapstructure:"lookback_seconds"`
	// BackfillEnabled: 是否允许全量回填
	BackfillEnabled bool `mapstructure:"backfill_enabled"`
	// BackfillMaxDays: 回填最大跨度（天）
	BackfillMaxDays int `mapstructure:"backfill_max_days"`
	// Retention: 各表保留窗口（天）
	Retention DashboardAggregationRetentionConfig `mapstructure:"retention"`
	// RecomputeDays: 启动时重算最近 N 天
	RecomputeDays int `mapstructure:"recompute_days"`
}

// DashboardAggregationRetentionConfig 预聚合保留窗口
type DashboardAggregationRetentionConfig struct {
	UsageLogsDays int `mapstructure:"usage_logs_days"`
	HourlyDays    int `mapstructure:"hourly_days"`
	DailyDays     int `mapstructure:"daily_days"`
}

// UsageCleanupConfig 使用记录清理任务配置
type UsageCleanupConfig struct {
	// Enabled: 是否启用清理任务执行器
	Enabled bool `mapstructure:"enabled"`
	// MaxRangeDays: 单次任务允许的最大时间跨度（天）
	MaxRangeDays int `mapstructure:"max_range_days"`
	// BatchSize: 单批删除数量
	BatchSize int `mapstructure:"batch_size"`
	// WorkerIntervalSeconds: 后台任务轮询间隔（秒）
	WorkerIntervalSeconds int `mapstructure:"worker_interval_seconds"`
	// TaskTimeoutSeconds: 单次任务最大执行时长（秒）
	TaskTimeoutSeconds int `mapstructure:"task_timeout_seconds"`
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
	// Optional: load env file defaults (does not override existing env vars).
	// This supports running the binary directly with deploy/.env(.example) without docker-compose env injection.
	loadEnvDefaultsFromFiles([]string{
		".env",
		filepath.Join("deploy", ".env"),
		".env.example",
		filepath.Join("deploy", ".env.example"),
	})

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths in priority order
	// 1. DATA_DIR environment variable (highest priority)
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		viper.AddConfigPath(dataDir)
	}
	// 2. Docker data directory
	viper.AddConfigPath("/app/data")
	// 3. Current directory
	viper.AddConfigPath(".")
	// 4. Config subdirectory
	viper.AddConfigPath("./config")
	// Workspace convenience: allow running the server from repo root.
	viper.AddConfigPath("./backend")
	viper.AddConfigPath("./backend/config")
	// Deployment convenience: allow using deploy/config.example.yaml as a base config.
	viper.AddConfigPath("./deploy")
	viper.AddConfigPath("/etc/sub2api")

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindCoreEnvAliases(viper.GetViper())
	bindLegacyEnvAliases(viper.GetViper())

	// 默认值
	setDefaults()

	// JSON env overrides for array-of-objects fields (docker-friendly).
	// These are used by deploy/.env.example and allow configuring complex arrays without editing YAML.
	if raw := strings.TrimSpace(os.Getenv("PAYMENT_PACKAGES")); raw != "" {
		var parsed []map[string]any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, fmt.Errorf("invalid PAYMENT_PACKAGES JSON: %w", err)
		}
		viper.Set("payment.packages", parsed)
	}
	if raw := strings.TrimSpace(os.Getenv("PROMOTION_TIERS")); raw != "" {
		var parsed []map[string]any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, fmt.Errorf("invalid PROMOTION_TIERS JSON: %w", err)
		}
		viper.Set("promotion.tiers", parsed)
	}

	// String-slice env overrides (docker-friendly). Accepts JSON array (["/a","/b"]) or CSV (/a,/b).
	if raw := strings.TrimSpace(os.Getenv("SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATHS")); raw != "" {
		paths, err := parseStringListEnv(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATHS: %w", err)
		}
		viper.Set("security.admin_api_key_read_only.allowed_paths", paths)
	}
	if raw := strings.TrimSpace(os.Getenv("SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATH_PREFIXES")); raw != "" {
		prefixes, err := parseStringListEnv(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATH_PREFIXES: %w", err)
		}
		viper.Set("security.admin_api_key_read_only.allowed_path_prefixes", prefixes)
	}

	// Int64-slice env overrides (docker-friendly). Accepts JSON array ([1,2] or ["1","2"]) or CSV (1,2).
	if raw := strings.TrimSpace(os.Getenv("GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_INCLUDE_ACCOUNT_IDS")); raw != "" {
		ids, err := parseInt64ListEnv(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_INCLUDE_ACCOUNT_IDS: %w", err)
		}
		viper.Set("gateway.scheduling.anthropic_apikey_monitor.include_account_ids", ids)
	}
	if raw := strings.TrimSpace(os.Getenv("GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_EXCLUDE_ACCOUNT_IDS")); raw != "" {
		ids, err := parseInt64ListEnv(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_EXCLUDE_ACCOUNT_IDS: %w", err)
		}
		viper.Set("gateway.scheduling.anthropic_apikey_monitor.exclude_account_ids", ids)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config error: %w", err)
		}
		// Fallback: allow using deploy/config.example.yaml when no config.yaml exists.
		if _, statErr := os.Stat(filepath.Join("deploy", "config.example.yaml")); statErr == nil {
			viper.SetConfigFile(filepath.Join("deploy", "config.example.yaml"))
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("read config error: %w", err)
			}
		}
		// 配置文件不存在时使用默认值
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	cfg.RunMode = NormalizeRunMode(cfg.RunMode)
	cfg.Server.Mode = strings.ToLower(strings.TrimSpace(cfg.Server.Mode))
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	cfg.JWT.Secret = strings.TrimSpace(cfg.JWT.Secret)
	cfg.CORS.AllowedOrigins = normalizeStringSlice(cfg.CORS.AllowedOrigins)
	cfg.Security.ResponseHeaders.AdditionalAllowed = normalizeStringSlice(cfg.Security.ResponseHeaders.AdditionalAllowed)
	cfg.Security.ResponseHeaders.ForceRemove = normalizeStringSlice(cfg.Security.ResponseHeaders.ForceRemove)
	cfg.Security.CSP.Policy = strings.TrimSpace(cfg.Security.CSP.Policy)
	cfg.Security.AdminAPIKeyReadOnly.AllowedPaths = normalizeStringSlice(cfg.Security.AdminAPIKeyReadOnly.AllowedPaths)
	cfg.Security.AdminAPIKeyReadOnly.AllowedPathPrefixes = normalizeStringSlice(cfg.Security.AdminAPIKeyReadOnly.AllowedPathPrefixes)
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.IncludeAccountIDs = normalizeInt64Slice(cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.IncludeAccountIDs)
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.ExcludeAccountIDs = normalizeInt64Slice(cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.ExcludeAccountIDs)

	if cfg.Server.Mode != "release" && cfg.JWT.Secret == "" {
		secret, err := generateJWTSecret(64)
		if err != nil {
			return nil, fmt.Errorf("generate jwt secret error: %w", err)
		}
		cfg.JWT.Secret = secret
		log.Println("Warning: JWT secret auto-generated for non-release mode. Do not use in production.")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	if !cfg.Security.URLAllowlist.Enabled {
		log.Println("Warning: security.url_allowlist.enabled=false; allowlist/SSRF checks disabled (minimal format validation only).")
	}
	if !cfg.Security.ResponseHeaders.Enabled {
		log.Println("Warning: security.response_headers.enabled=false; configurable header filtering disabled (default allowlist only).")
	}

	if cfg.Server.Mode != "release" && cfg.JWT.Secret != "" && isWeakJWTSecret(cfg.JWT.Secret) {
		log.Println("Warning: JWT secret appears weak; use a 32+ character random secret in production.")
	}
	if len(cfg.Security.ResponseHeaders.AdditionalAllowed) > 0 || len(cfg.Security.ResponseHeaders.ForceRemove) > 0 {
		log.Printf("AUDIT: response header policy configured additional_allowed=%v force_remove=%v",
			cfg.Security.ResponseHeaders.AdditionalAllowed,
			cfg.Security.ResponseHeaders.ForceRemove,
		)
	}

	return &cfg, nil
}

func loadEnvDefaultsFromFiles(paths []string) {
	for _, path := range paths {
		loadEnvDefaultsFromFile(path)
	}
}

// loadEnvDefaultsFromFile loads KEY=VALUE pairs from file into the process environment
// only when the key is not already set.
func loadEnvDefaultsFromFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		_ = os.Setenv(key, value)
	}
}

// bindCoreEnvAliases binds docker-compose / setup env vars to the config keys.
//
// Viper's AutomaticEnv is not always sufficient for nested structs when using Unmarshal,
// so we bind the common env vars explicitly to guarantee env overrides are applied.
func bindCoreEnvAliases(v *viper.Viper) {
	if v == nil {
		return
	}

	// Server
	_ = v.BindEnv("server.host", "SERVER_HOST")
	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("server.mode", "SERVER_MODE", "GIN_MODE")
	_ = v.BindEnv("run_mode", "RUN_MODE")
	_ = v.BindEnv("timezone", "TIMEZONE", "TZ")

	// Database
	_ = v.BindEnv("database.host", "DATABASE_HOST")
	_ = v.BindEnv("database.port", "DATABASE_PORT")
	_ = v.BindEnv("database.user", "DATABASE_USER")
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = v.BindEnv("database.dbname", "DATABASE_DBNAME")
	_ = v.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	_ = v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime_minutes", "DATABASE_CONN_MAX_LIFETIME_MINUTES")
	_ = v.BindEnv("database.conn_max_idle_time_minutes", "DATABASE_CONN_MAX_IDLE_TIME_MINUTES")

	// Redis
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("redis.dial_timeout_seconds", "REDIS_DIAL_TIMEOUT_SECONDS")
	_ = v.BindEnv("redis.read_timeout_seconds", "REDIS_READ_TIMEOUT_SECONDS")
	_ = v.BindEnv("redis.write_timeout_seconds", "REDIS_WRITE_TIMEOUT_SECONDS")
	_ = v.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	_ = v.BindEnv("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")

	// JWT
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.expire_hour", "JWT_EXPIRE_HOUR")

	// Rate limiting
	_ = v.BindEnv("rate_limit.overload_cooldown_minutes", "RATE_LIMIT_OVERLOAD_COOLDOWN_MINUTES")
	_ = v.BindEnv("rate_limit.fallback_cooldown_minutes", "RATE_LIMIT_FALLBACK_COOLDOWN_MINUTES")

	// Security (admin read-only API key allowlist)
	_ = v.BindEnv("security.admin_api_key_read_only.allowed_paths", "SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATHS")
	_ = v.BindEnv("security.admin_api_key_read_only.allowed_path_prefixes", "SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATH_PREFIXES")

	// Dashboard aggregation / retention
	_ = v.BindEnv("dashboard_aggregation.enabled", "DASHBOARD_AGGREGATION_ENABLED")
	_ = v.BindEnv("dashboard_aggregation.interval_seconds", "DASHBOARD_AGGREGATION_INTERVAL_SECONDS")
	_ = v.BindEnv("dashboard_aggregation.lookback_seconds", "DASHBOARD_AGGREGATION_LOOKBACK_SECONDS")
	_ = v.BindEnv("dashboard_aggregation.backfill_enabled", "DASHBOARD_AGGREGATION_BACKFILL_ENABLED")
	_ = v.BindEnv("dashboard_aggregation.backfill_max_days", "DASHBOARD_AGGREGATION_BACKFILL_MAX_DAYS")
	_ = v.BindEnv("dashboard_aggregation.recompute_days", "DASHBOARD_AGGREGATION_RECOMPUTE_DAYS")
	_ = v.BindEnv("dashboard_aggregation.retention.usage_logs_days", "DASHBOARD_AGGREGATION_RETENTION_USAGE_LOGS_DAYS")
	_ = v.BindEnv("dashboard_aggregation.retention.hourly_days", "DASHBOARD_AGGREGATION_RETENTION_HOURLY_DAYS")
	_ = v.BindEnv("dashboard_aggregation.retention.daily_days", "DASHBOARD_AGGREGATION_RETENTION_DAILY_DAYS")

	// Usage cleanup task
	_ = v.BindEnv("usage_cleanup.enabled", "USAGE_CLEANUP_ENABLED")
	_ = v.BindEnv("usage_cleanup.max_range_days", "USAGE_CLEANUP_MAX_RANGE_DAYS")
	_ = v.BindEnv("usage_cleanup.batch_size", "USAGE_CLEANUP_BATCH_SIZE")
	_ = v.BindEnv("usage_cleanup.worker_interval_seconds", "USAGE_CLEANUP_WORKER_INTERVAL_SECONDS")
	_ = v.BindEnv("usage_cleanup.task_timeout_seconds", "USAGE_CLEANUP_TASK_TIMEOUT_SECONDS")

	// Gateway scheduling monitor (Anthropic API-key health)
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.enabled", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_ENABLED")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.interval", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_INTERVAL")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.failure_threshold", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_FAILURE_THRESHOLD")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.success_threshold", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_SUCCESS_THRESHOLD")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.request_timeout", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_REQUEST_TIMEOUT")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.max_concurrency", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_MAX_CONCURRENCY")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.model_id", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_MODEL_ID")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.include_account_ids", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_INCLUDE_ACCOUNT_IDS")
	_ = v.BindEnv("gateway.scheduling.anthropic_apikey_monitor.exclude_account_ids", "GATEWAY_SCHEDULING_ANTHROPIC_APIKEY_MONITOR_EXCLUDE_ACCOUNT_IDS")

	// Gateway TLS fingerprint simulation
	_ = v.BindEnv("gateway.tls_fingerprint.enabled", "GATEWAY_TLS_FINGERPRINT_ENABLED")

	// Gemini OAuth / Quota
	_ = v.BindEnv("gemini.oauth.client_id", "GEMINI_OAUTH_CLIENT_ID")
	_ = v.BindEnv("gemini.oauth.client_secret", "GEMINI_OAUTH_CLIENT_SECRET")
	_ = v.BindEnv("gemini.oauth.scopes", "GEMINI_OAUTH_SCOPES")
	_ = v.BindEnv("gemini.quota.policy", "GEMINI_QUOTA_POLICY")

	// Payment
	_ = v.BindEnv("payment.enabled", "PAYMENT_ENABLED")
	_ = v.BindEnv("payment.base_url", "PAYMENT_BASE_URL")
	_ = v.BindEnv("payment.min_amount", "PAYMENT_MIN_AMOUNT")
	_ = v.BindEnv("payment.max_amount", "PAYMENT_MAX_AMOUNT")
	_ = v.BindEnv("payment.exchange_rate", "PAYMENT_EXCHANGE_RATE")
	_ = v.BindEnv("payment.discount_rate", "PAYMENT_DISCOUNT_RATE")
	_ = v.BindEnv("payment.order_expire_minutes", "PAYMENT_ORDER_EXPIRE_MINUTES")
	_ = v.BindEnv("payment.max_orders_per_minute", "PAYMENT_MAX_ORDERS_PER_MINUTE")
	_ = v.BindEnv("payment.order_prefix", "PAYMENT_ORDER_PREFIX")
	// Complex arrays are also supported via JSON env overrides in Load(): PAYMENT_PACKAGES.

	// Payment.ZPay
	_ = v.BindEnv("payment.zpay.enabled", "PAYMENT_ZPAY_ENABLED", "ZPAY_ENABLED")
	_ = v.BindEnv("payment.zpay.pid", "PAYMENT_ZPAY_PID", "ZPAY_PID")
	_ = v.BindEnv("payment.zpay.key", "PAYMENT_ZPAY_KEY", "ZPAY_KEY")
	_ = v.BindEnv("payment.zpay.api_url", "PAYMENT_ZPAY_API_URL", "ZPAY_API_URL")
	_ = v.BindEnv("payment.zpay.submit_url", "PAYMENT_ZPAY_SUBMIT_URL", "ZPAY_SUBMIT_URL")
	_ = v.BindEnv("payment.zpay.query_url", "PAYMENT_ZPAY_QUERY_URL", "ZPAY_QUERY_URL")
	_ = v.BindEnv("payment.zpay.payment_methods", "PAYMENT_ZPAY_PAYMENT_METHODS", "ZPAY_PAYMENT_METHODS")
	_ = v.BindEnv("payment.zpay.order_prefix", "PAYMENT_ZPAY_ORDER_PREFIX", "ZPAY_ORDER_PREFIX")
	_ = v.BindEnv("payment.zpay.notify_url", "PAYMENT_ZPAY_NOTIFY_URL", "ZPAY_NOTIFY_URL")
	_ = v.BindEnv("payment.zpay.return_url", "PAYMENT_ZPAY_RETURN_URL", "ZPAY_RETURN_URL")
	_ = v.BindEnv("payment.zpay.notify_user", "PAYMENT_ZPAY_NOTIFY_USER", "ZPAY_NOTIFY_USER")
	_ = v.BindEnv("payment.zpay.ip_whitelist", "PAYMENT_ZPAY_IP_WHITELIST", "ZPAY_IP_WHITELIST")
	_ = v.BindEnv("payment.zpay.require_https", "PAYMENT_ZPAY_REQUIRE_HTTPS", "ZPAY_REQUIRE_HTTPS")
	// ZPay multi-channel support: channel ID (cid) for specific payment types
	_ = v.BindEnv("payment.zpay.alipay_channel_id", "PAYMENT_ZPAY_ALIPAY_CHANNEL_ID", "ZPAY_ALIPAY_CHANNEL_ID")
	_ = v.BindEnv("payment.zpay.wechat_channel_id", "PAYMENT_ZPAY_WECHAT_CHANNEL_ID", "ZPAY_WECHAT_CHANNEL_ID")

	// Payment.Stripe (keep deploy/.env.example naming working)
	_ = v.BindEnv("payment.stripe.enabled", "PAYMENT_STRIPE_ENABLED", "STRIPE_ENABLED")
	_ = v.BindEnv("payment.stripe.api_key", "PAYMENT_STRIPE_API_KEY", "STRIPE_API_KEY", "STRIPE_SECRET_KEY")
	_ = v.BindEnv("payment.stripe.webhook_secret", "PAYMENT_STRIPE_WEBHOOK_SECRET", "STRIPE_WEBHOOK_SECRET")
	_ = v.BindEnv("payment.stripe.api_version", "PAYMENT_STRIPE_API_VERSION", "STRIPE_API_VERSION")
	_ = v.BindEnv("payment.stripe.payment_methods", "PAYMENT_STRIPE_PAYMENT_METHODS", "STRIPE_PAYMENT_METHODS")
	_ = v.BindEnv("payment.stripe.currency", "PAYMENT_STRIPE_CURRENCY", "STRIPE_CURRENCY")
	_ = v.BindEnv("payment.stripe.success_url", "PAYMENT_STRIPE_SUCCESS_URL", "STRIPE_SUCCESS_URL")
	_ = v.BindEnv("payment.stripe.cancel_url", "PAYMENT_STRIPE_CANCEL_URL", "STRIPE_CANCEL_URL")
	_ = v.BindEnv("payment.stripe.wechat_client", "PAYMENT_STRIPE_WECHAT_CLIENT", "STRIPE_WECHAT_CLIENT")
	_ = v.BindEnv("payment.stripe.wechat_app_id", "PAYMENT_STRIPE_WECHAT_APP_ID", "STRIPE_WECHAT_APP_ID")
}

// bindLegacyEnvAliases binds legacy (non-namespaced) env vars to the current config keys.
//
// This repo primarily uses Viper's `.` -> `_` mapping, so config keys like:
//
//	payment.zpay.enabled -> PAYMENT_ZPAY_ENABLED
//
// But some deployments already use:
//
//	ZPAY_ENABLED / STRIPE_SECRET_KEY / ...
//
// This helper keeps backwards compatibility.
func bindLegacyEnvAliases(v *viper.Viper) {
	if v == nil {
		return
	}

	// Server
	// Many deployments use Gin's standard env var; keep it working.
	_ = v.BindEnv("server.mode", "SERVER_MODE", "GIN_MODE")

	// Payment
	_ = v.BindEnv("payment.order_prefix", "PAYMENT_ORDER_PREFIX")

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
	// ZPay multi-channel support (legacy env vars)
	_ = v.BindEnv("payment.zpay.alipay_channel_id", "ZPAY_ALIPAY_CHANNEL_ID")
	_ = v.BindEnv("payment.zpay.wechat_channel_id", "ZPAY_WECHAT_CHANNEL_ID")

	// Stripe
	_ = v.BindEnv("payment.stripe.enabled", "STRIPE_ENABLED", "PAYMENT_STRIPE_ENABLED")
	// Common naming: STRIPE_SECRET_KEY is the Stripe API key.
	_ = v.BindEnv("payment.stripe.api_key", "STRIPE_API_KEY", "STRIPE_SECRET_KEY", "PAYMENT_STRIPE_API_KEY")
	_ = v.BindEnv("payment.stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET", "PAYMENT_STRIPE_WEBHOOK_SECRET")
	_ = v.BindEnv("payment.stripe.api_version", "STRIPE_API_VERSION", "PAYMENT_STRIPE_API_VERSION")
	_ = v.BindEnv("payment.stripe.payment_methods", "STRIPE_PAYMENT_METHODS", "PAYMENT_STRIPE_PAYMENT_METHODS")
	_ = v.BindEnv("payment.stripe.currency", "STRIPE_CURRENCY", "PAYMENT_STRIPE_CURRENCY")
	_ = v.BindEnv("payment.stripe.success_url", "STRIPE_SUCCESS_URL", "PAYMENT_STRIPE_SUCCESS_URL")
	_ = v.BindEnv("payment.stripe.cancel_url", "STRIPE_CANCEL_URL", "PAYMENT_STRIPE_CANCEL_URL")
	_ = v.BindEnv("payment.stripe.wechat_client", "STRIPE_WECHAT_CLIENT", "PAYMENT_STRIPE_WECHAT_CLIENT")
	_ = v.BindEnv("payment.stripe.wechat_app_id", "STRIPE_WECHAT_APP_ID", "PAYMENT_STRIPE_WECHAT_APP_ID")

	// Dingtalk
	_ = v.BindEnv("dingtalk.enabled", "DINGTALK_ENABLED")
	_ = v.BindEnv("dingtalk.env", "DINGTALK_ENV", "ALERT_ENV", "APP_ENV")
	_ = v.BindEnv("dingtalk.webhook_url", "DINGTALK_WEBHOOK_URL")
	_ = v.BindEnv("dingtalk.secret", "DINGTALK_SECRET")
	_ = v.BindEnv("dingtalk.at_mobiles", "DINGTALK_AT_MOBILES")
	_ = v.BindEnv("dingtalk.at_all", "DINGTALK_AT_ALL")
	_ = v.BindEnv("dingtalk.payment_notify_enabled", "DINGTALK_PAYMENT_NOTIFY_ENABLED")
	// Dingtalk bot (incoming)
	_ = v.BindEnv("dingtalk_bot.enabled", "DINGTALK_BOT_ENABLED")
	_ = v.BindEnv("dingtalk_bot.access_token", "DINGTALK_BOT_ACCESS_TOKEN")
	_ = v.BindEnv("dingtalk_bot.sign_secret", "DINGTALK_BOT_SIGN_SECRET")
	_ = v.BindEnv("dingtalk_bot.allowed_sender_ids", "DINGTALK_BOT_ALLOWED_SENDER_IDS")
	_ = v.BindEnv("dingtalk_bot.default_remark", "DINGTALK_BOT_DEFAULT_REMARK")

	// Referral (legacy and docker env friendly)
	_ = v.BindEnv("referral.enabled", "REFERRAL_ENABLED")
	_ = v.BindEnv("referral.base_url", "REFERRAL_BASE_URL")
	_ = v.BindEnv("referral.link_base_url", "REFERRAL_LINK_BASE_URL")
	_ = v.BindEnv("referral.reward_usd", "REFERRAL_REWARD_USD")
	_ = v.BindEnv("referral.qualified_recharge_cny", "REFERRAL_QUALIFIED_RECHARGE_CNY")
	_ = v.BindEnv("referral.qualified_recharge_usd", "REFERRAL_QUALIFIED_RECHARGE_USD")
	_ = v.BindEnv("referral.code_length", "REFERRAL_CODE_LENGTH")
	_ = v.BindEnv("referral.max_invitees_per_user", "REFERRAL_MAX_INVITEES_PER_USER")
}

func setDefaults() {
	viper.SetDefault("run_mode", RunModeStandard)

	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_header_timeout", 30) // 30秒读取请求头
	viper.SetDefault("server.idle_timeout", 120)       // 120秒空闲超时
	viper.SetDefault("server.trusted_proxies", []string{})

	// CORS
	viper.SetDefault("cors.allowed_origins", []string{})
	viper.SetDefault("cors.allow_credentials", true)

	// Security
	viper.SetDefault("security.url_allowlist.enabled", false)
	viper.SetDefault("security.url_allowlist.upstream_hosts", []string{
		"api.openai.com",
		"api.anthropic.com",
		"api.kimi.com",
		"open.bigmodel.cn",
		"api.minimaxi.com",
		"generativelanguage.googleapis.com",
		"cloudcode-pa.googleapis.com",
		"*.openai.azure.com",
	})
	viper.SetDefault("security.url_allowlist.pricing_hosts", []string{
		"raw.githubusercontent.com",
	})
	viper.SetDefault("security.url_allowlist.crs_hosts", []string{})
	viper.SetDefault("security.url_allowlist.allow_private_hosts", true)
	viper.SetDefault("security.url_allowlist.allow_insecure_http", true)
	viper.SetDefault("security.response_headers.enabled", false)
	viper.SetDefault("security.response_headers.additional_allowed", []string{})
	viper.SetDefault("security.response_headers.force_remove", []string{})
	viper.SetDefault("security.csp.enabled", true)
	viper.SetDefault("security.csp.policy", DefaultCSPPolicy)
	viper.SetDefault("security.proxy_probe.insecure_skip_verify", false)
	// Read-only Admin API key allowlist (only affects x-api-key admin-ro-* auth path)
	// NOTE: The middleware also enforces method=GET for the read-only key.
	viper.SetDefault("security.admin_api_key_read_only.allowed_paths", []string{
		"/api/v1/admin/users/export",
		"/api/v1/admin/usage",
		"/api/v1/admin/payment/orders/export",
	})
	viper.SetDefault("security.admin_api_key_read_only.allowed_path_prefixes", []string{})

	// Billing
	viper.SetDefault("billing.circuit_breaker.enabled", true)
	viper.SetDefault("billing.circuit_breaker.failure_threshold", 5)
	viper.SetDefault("billing.circuit_breaker.reset_timeout_seconds", 30)
	viper.SetDefault("billing.circuit_breaker.half_open_requests", 3)

	// Turnstile
	viper.SetDefault("turnstile.required", false)

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "sub2api")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 50)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime_minutes", 30)
	viper.SetDefault("database.conn_max_idle_time_minutes", 5)

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.dial_timeout_seconds", 5)
	viper.SetDefault("redis.read_timeout_seconds", 3)
	viper.SetDefault("redis.write_timeout_seconds", 3)
	viper.SetDefault("redis.pool_size", 128)
	viper.SetDefault("redis.min_idle_conns", 10)

	// Ops (vNext)
	viper.SetDefault("ops.enabled", true)
	viper.SetDefault("ops.use_preaggregated_tables", false)
	viper.SetDefault("ops.cleanup.enabled", true)
	viper.SetDefault("ops.cleanup.schedule", "0 2 * * *")
	// Retention days: vNext defaults to 30 days across ops datasets.
	viper.SetDefault("ops.cleanup.error_log_retention_days", 30)
	viper.SetDefault("ops.cleanup.minute_metrics_retention_days", 30)
	viper.SetDefault("ops.cleanup.hourly_metrics_retention_days", 30)
	viper.SetDefault("ops.aggregation.enabled", true)
	viper.SetDefault("ops.metrics_collector_cache.enabled", true)
	// TTL should be slightly larger than collection interval (1m) to maximize cross-replica cache hits.
	viper.SetDefault("ops.metrics_collector_cache.ttl", 65*time.Second)

	// JWT
	viper.SetDefault("jwt.secret", "")
	viper.SetDefault("jwt.expire_hour", 24)

	// Default
	// Admin credentials are created via the setup flow (web wizard / CLI / AUTO_SETUP).
	// Do not ship fixed defaults here to avoid insecure "known credentials" in production.
	viper.SetDefault("default.admin_email", "")
	viper.SetDefault("default.admin_password", "")
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
	// Default to the SPA register route; ReferralHandler will convert to absolute if payment.base_url is set.
	viper.SetDefault("referral.link_base_url", "/register")
	viper.SetDefault("referral.base_url", "")
	viper.SetDefault("referral.reward_usd", 0.0)
	viper.SetDefault("referral.qualified_recharge_cny", 0.0)
	viper.SetDefault("referral.qualified_recharge_usd", 0.0)
	// Optional: if empty, all recharge types count toward qualification.
	viper.SetDefault("referral.qualified_recharge_types", []string{})
	viper.SetDefault("referral.code_length", 8)
	viper.SetDefault("referral.max_invitees_per_user", 0)

	// Dingtalk
	viper.SetDefault("dingtalk.enabled", false)
	viper.SetDefault("dingtalk.env", "")
	viper.SetDefault("dingtalk.webhook_url", "")
	viper.SetDefault("dingtalk.secret", "")
	viper.SetDefault("dingtalk.at_mobiles", "")
	viper.SetDefault("dingtalk.at_all", false)
	viper.SetDefault("dingtalk.payment_notify_enabled", true)
	// Dingtalk bot (incoming)
	viper.SetDefault("dingtalk_bot.enabled", false)
	viper.SetDefault("dingtalk_bot.access_token", "")
	viper.SetDefault("dingtalk_bot.sign_secret", "")
	viper.SetDefault("dingtalk_bot.allowed_sender_ids", "")
	viper.SetDefault("dingtalk_bot.default_remark", "")

	// Payment
	viper.SetDefault("payment.enabled", false)
	viper.SetDefault("payment.base_url", "")
	viper.SetDefault("payment.min_amount", 1.0)
	viper.SetDefault("payment.max_amount", 10000.0)
	viper.SetDefault("payment.exchange_rate", 7.2)
	viper.SetDefault("payment.discount_rate", 1.0)
	viper.SetDefault("payment.order_expire_minutes", 30)
	viper.SetDefault("payment.max_orders_per_minute", 3)
	viper.SetDefault("payment.order_prefix", "PO")
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
	viper.SetDefault("rate_limit.fallback_cooldown_minutes", 5)

	// Pricing - 从 price-mirror 分支同步，该分支维护了 sha256 哈希文件用于增量更新检查
	viper.SetDefault("pricing.remote_url", "https://raw.githubusercontent.com/Wei-Shaw/claude-relay-service/price-mirror/model_prices_and_context_window.json")
	viper.SetDefault("pricing.hash_url", "https://raw.githubusercontent.com/Wei-Shaw/claude-relay-service/price-mirror/model_prices_and_context_window.sha256")
	viper.SetDefault("pricing.data_dir", "./data")
	viper.SetDefault("pricing.fallback_file", "./resources/model-pricing/model_prices_and_context_window.json")
	viper.SetDefault("pricing.update_interval_hours", 24)
	viper.SetDefault("pricing.hash_check_interval_minutes", 10)

	// Timezone (default to Asia/Shanghai for Chinese users)
	viper.SetDefault("timezone", "Asia/Shanghai")

	// API Key auth cache
	viper.SetDefault("api_key_auth_cache.l1_size", 65535)
	viper.SetDefault("api_key_auth_cache.l1_ttl_seconds", 15)
	viper.SetDefault("api_key_auth_cache.l2_ttl_seconds", 300)
	viper.SetDefault("api_key_auth_cache.negative_ttl_seconds", 30)
	viper.SetDefault("api_key_auth_cache.jitter_percent", 10)
	viper.SetDefault("api_key_auth_cache.singleflight", true)

	// Dashboard cache
	viper.SetDefault("dashboard_cache.enabled", true)
	viper.SetDefault("dashboard_cache.key_prefix", "sub2api:")
	viper.SetDefault("dashboard_cache.stats_fresh_ttl_seconds", 15)
	viper.SetDefault("dashboard_cache.stats_ttl_seconds", 30)
	viper.SetDefault("dashboard_cache.stats_refresh_timeout_seconds", 30)

	// Dashboard aggregation
	viper.SetDefault("dashboard_aggregation.enabled", true)
	viper.SetDefault("dashboard_aggregation.interval_seconds", 60)
	viper.SetDefault("dashboard_aggregation.lookback_seconds", 120)
	viper.SetDefault("dashboard_aggregation.backfill_enabled", false)
	viper.SetDefault("dashboard_aggregation.backfill_max_days", 31)
	viper.SetDefault("dashboard_aggregation.retention.usage_logs_days", 90)
	viper.SetDefault("dashboard_aggregation.retention.hourly_days", 180)
	viper.SetDefault("dashboard_aggregation.retention.daily_days", 730)
	viper.SetDefault("dashboard_aggregation.recompute_days", 2)

	// Usage cleanup task
	viper.SetDefault("usage_cleanup.enabled", true)
	viper.SetDefault("usage_cleanup.max_range_days", 31)
	viper.SetDefault("usage_cleanup.batch_size", 5000)
	viper.SetDefault("usage_cleanup.worker_interval_seconds", 10)
	viper.SetDefault("usage_cleanup.task_timeout_seconds", 1800)

	// Gateway
	viper.SetDefault("gateway.response_header_timeout", 600) // 600秒(10分钟)等待上游响应头，LLM高负载时可能排队较久
	viper.SetDefault("gateway.log_upstream_error_body", false)
	viper.SetDefault("gateway.log_upstream_error_body_max_bytes", 2048)
	viper.SetDefault("gateway.inject_beta_for_apikey", false)
	viper.SetDefault("gateway.failover_on_400", false)
	viper.SetDefault("gateway.max_account_switches", 10)
	viper.SetDefault("gateway.max_account_switches_gemini", 3)
	viper.SetDefault("gateway.antigravity_fallback_cooldown_minutes", 1)
	viper.SetDefault("gateway.max_body_size", int64(100*1024*1024))
	viper.SetDefault("gateway.connection_pool_isolation", ConnectionPoolIsolationAccountProxy)
	// HTTP 上游连接池配置（针对 5000+ 并发用户优化）
	viper.SetDefault("gateway.max_idle_conns", 240)           // 最大空闲连接总数（HTTP/2 场景默认）
	viper.SetDefault("gateway.max_idle_conns_per_host", 120)  // 每主机最大空闲连接（HTTP/2 场景默认）
	viper.SetDefault("gateway.max_conns_per_host", 240)       // 每主机最大连接数（含活跃，HTTP/2 场景默认）
	viper.SetDefault("gateway.idle_conn_timeout_seconds", 90) // 空闲连接超时（秒）
	viper.SetDefault("gateway.max_upstream_clients", 5000)
	viper.SetDefault("gateway.client_idle_ttl_seconds", 900)
	viper.SetDefault("gateway.concurrency_slot_ttl_minutes", 30) // 并发槽位过期时间（支持超长请求）
	viper.SetDefault("gateway.stream_data_interval_timeout", 180)
	viper.SetDefault("gateway.stream_keepalive_interval", 10)
	viper.SetDefault("gateway.max_line_size", 40*1024*1024)
	viper.SetDefault("gateway.scheduling.sticky_session_max_waiting", 3)
	viper.SetDefault("gateway.scheduling.sticky_session_wait_timeout", 120*time.Second)
	viper.SetDefault("gateway.scheduling.fallback_wait_timeout", 30*time.Second)
	viper.SetDefault("gateway.scheduling.fallback_max_waiting", 100)
	viper.SetDefault("gateway.scheduling.fallback_selection_mode", "last_used")
	viper.SetDefault("gateway.scheduling.load_batch_enabled", true)
	viper.SetDefault("gateway.scheduling.slot_cleanup_interval", 30*time.Second)
	viper.SetDefault("gateway.scheduling.db_fallback_enabled", true)
	viper.SetDefault("gateway.scheduling.db_fallback_timeout_seconds", 0)
	viper.SetDefault("gateway.scheduling.db_fallback_max_qps", 0)
	viper.SetDefault("gateway.scheduling.outbox_poll_interval_seconds", 1)
	viper.SetDefault("gateway.scheduling.outbox_lag_warn_seconds", 5)
	viper.SetDefault("gateway.scheduling.outbox_lag_rebuild_seconds", 10)
	viper.SetDefault("gateway.scheduling.outbox_lag_rebuild_failures", 3)
	viper.SetDefault("gateway.scheduling.outbox_backlog_rebuild_rows", 10000)
	viper.SetDefault("gateway.scheduling.full_rebuild_interval_seconds", 300)
	// Anthropic API-key connectivity monitor (disabled by default)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.enabled", false)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.interval", 10*time.Second)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.failure_threshold", 6)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.success_threshold", 6)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.request_timeout", 8*time.Second)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.max_concurrency", 4)
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.model_id", "")
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.include_account_ids", []int64{})
	viper.SetDefault("gateway.scheduling.anthropic_apikey_monitor.exclude_account_ids", []int64{})
	// TLS指纹伪装配置（默认关闭，需要账号级别单独启用）
	viper.SetDefault("gateway.tls_fingerprint.enabled", true)
	viper.SetDefault("concurrency.ping_interval", 10)

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
	viper.SetDefault("gemini.quota.policy", "")
}

func (c *Config) Validate() error {
	if c.Server.Mode == "release" {
		if c.JWT.Secret == "" {
			return fmt.Errorf("jwt.secret is required in release mode")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be at least 32 characters")
		}
		if isWeakJWTSecret(c.JWT.Secret) {
			return fmt.Errorf("jwt.secret is too weak")
		}
	}
	if c.JWT.ExpireHour <= 0 {
		return fmt.Errorf("jwt.expire_hour must be positive")
	}
	if c.JWT.ExpireHour > 168 {
		return fmt.Errorf("jwt.expire_hour must be <= 168 (7 days)")
	}
	if c.JWT.ExpireHour > 24 {
		log.Printf("Warning: jwt.expire_hour is %d hours (> 24). Consider shorter expiration for security.", c.JWT.ExpireHour)
	}
	if c.Security.CSP.Enabled && strings.TrimSpace(c.Security.CSP.Policy) == "" {
		return fmt.Errorf("security.csp.policy is required when CSP is enabled")
	}
	for i, raw := range c.Security.AdminAPIKeyReadOnly.AllowedPaths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		if strings.ContainsAny(path, "\r\n") {
			return fmt.Errorf("security.admin_api_key_read_only.allowed_paths[%d] contains invalid characters", i)
		}
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("security.admin_api_key_read_only.allowed_paths[%d] must start with '/'", i)
		}
	}
	for i, raw := range c.Security.AdminAPIKeyReadOnly.AllowedPathPrefixes {
		prefix := strings.TrimSpace(raw)
		if prefix == "" {
			continue
		}
		if strings.ContainsAny(prefix, "\r\n") {
			return fmt.Errorf("security.admin_api_key_read_only.allowed_path_prefixes[%d] contains invalid characters", i)
		}
		if !strings.HasPrefix(prefix, "/") {
			return fmt.Errorf("security.admin_api_key_read_only.allowed_path_prefixes[%d] must start with '/'", i)
		}
	}
	if c.LinuxDo.Enabled {
		if strings.TrimSpace(c.LinuxDo.ClientID) == "" {
			return fmt.Errorf("linuxdo_connect.client_id is required")
		}
		if err := ValidateFrontendRedirectURL(c.LinuxDo.FrontendRedirectURL); err != nil {
			return fmt.Errorf("linuxdo_connect.frontend_redirect_url: %w", err)
		}
		switch strings.TrimSpace(c.LinuxDo.TokenAuthMethod) {
		case "client_secret_post", "client_secret_basic", "none":
		default:
			return fmt.Errorf("linuxdo_connect.token_auth_method is invalid")
		}
		if strings.TrimSpace(c.LinuxDo.TokenAuthMethod) == "none" && !c.LinuxDo.UsePKCE {
			return fmt.Errorf("linuxdo_connect.use_pkce is required when token_auth_method=none")
		}
	}
	if c.Billing.CircuitBreaker.Enabled {
		if c.Billing.CircuitBreaker.FailureThreshold <= 0 {
			return fmt.Errorf("billing.circuit_breaker.failure_threshold must be positive")
		}
		if c.Billing.CircuitBreaker.ResetTimeoutSeconds <= 0 {
			return fmt.Errorf("billing.circuit_breaker.reset_timeout_seconds must be positive")
		}
		if c.Billing.CircuitBreaker.HalfOpenRequests <= 0 {
			return fmt.Errorf("billing.circuit_breaker.half_open_requests must be positive")
		}
	}
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database.max_open_conns must be positive")
	}
	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database.max_idle_conns must be non-negative")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database.max_idle_conns cannot exceed database.max_open_conns")
	}
	if c.Database.ConnMaxLifetimeMinutes < 0 {
		return fmt.Errorf("database.conn_max_lifetime_minutes must be non-negative")
	}
	if c.Database.ConnMaxIdleTimeMinutes < 0 {
		return fmt.Errorf("database.conn_max_idle_time_minutes must be non-negative")
	}
	if c.Redis.DialTimeoutSeconds <= 0 {
		return fmt.Errorf("redis.dial_timeout_seconds must be positive")
	}
	if c.Redis.ReadTimeoutSeconds <= 0 {
		return fmt.Errorf("redis.read_timeout_seconds must be positive")
	}
	if c.Redis.WriteTimeoutSeconds <= 0 {
		return fmt.Errorf("redis.write_timeout_seconds must be positive")
	}
	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis.pool_size must be positive")
	}
	if c.Redis.MinIdleConns < 0 {
		return fmt.Errorf("redis.min_idle_conns must be non-negative")
	}
	if c.Redis.MinIdleConns > c.Redis.PoolSize {
		return fmt.Errorf("redis.min_idle_conns cannot exceed redis.pool_size")
	}
	if c.Dashboard.Enabled {
		if c.Dashboard.StatsFreshTTLSeconds <= 0 {
			return fmt.Errorf("dashboard_cache.stats_fresh_ttl_seconds must be positive")
		}
		if c.Dashboard.StatsTTLSeconds <= 0 {
			return fmt.Errorf("dashboard_cache.stats_ttl_seconds must be positive")
		}
		if c.Dashboard.StatsRefreshTimeoutSeconds <= 0 {
			return fmt.Errorf("dashboard_cache.stats_refresh_timeout_seconds must be positive")
		}
		if c.Dashboard.StatsFreshTTLSeconds > c.Dashboard.StatsTTLSeconds {
			return fmt.Errorf("dashboard_cache.stats_fresh_ttl_seconds must be <= dashboard_cache.stats_ttl_seconds")
		}
	} else {
		if c.Dashboard.StatsFreshTTLSeconds < 0 {
			return fmt.Errorf("dashboard_cache.stats_fresh_ttl_seconds must be non-negative")
		}
		if c.Dashboard.StatsTTLSeconds < 0 {
			return fmt.Errorf("dashboard_cache.stats_ttl_seconds must be non-negative")
		}
		if c.Dashboard.StatsRefreshTimeoutSeconds < 0 {
			return fmt.Errorf("dashboard_cache.stats_refresh_timeout_seconds must be non-negative")
		}
	}
	if c.DashboardAgg.Enabled {
		if c.DashboardAgg.IntervalSeconds <= 0 {
			return fmt.Errorf("dashboard_aggregation.interval_seconds must be positive")
		}
		if c.DashboardAgg.LookbackSeconds < 0 {
			return fmt.Errorf("dashboard_aggregation.lookback_seconds must be non-negative")
		}
		if c.DashboardAgg.BackfillMaxDays < 0 {
			return fmt.Errorf("dashboard_aggregation.backfill_max_days must be non-negative")
		}
		if c.DashboardAgg.BackfillEnabled && c.DashboardAgg.BackfillMaxDays == 0 {
			return fmt.Errorf("dashboard_aggregation.backfill_max_days must be positive")
		}
		if c.DashboardAgg.Retention.UsageLogsDays <= 0 {
			return fmt.Errorf("dashboard_aggregation.retention.usage_logs_days must be positive")
		}
		if c.DashboardAgg.Retention.HourlyDays <= 0 {
			return fmt.Errorf("dashboard_aggregation.retention.hourly_days must be positive")
		}
		if c.DashboardAgg.Retention.DailyDays <= 0 {
			return fmt.Errorf("dashboard_aggregation.retention.daily_days must be positive")
		}
		if c.DashboardAgg.RecomputeDays < 0 {
			return fmt.Errorf("dashboard_aggregation.recompute_days must be non-negative")
		}
	} else {
		if c.DashboardAgg.IntervalSeconds < 0 {
			return fmt.Errorf("dashboard_aggregation.interval_seconds must be non-negative")
		}
		if c.DashboardAgg.LookbackSeconds < 0 {
			return fmt.Errorf("dashboard_aggregation.lookback_seconds must be non-negative")
		}
		if c.DashboardAgg.BackfillMaxDays < 0 {
			return fmt.Errorf("dashboard_aggregation.backfill_max_days must be non-negative")
		}
		if c.DashboardAgg.Retention.UsageLogsDays < 0 {
			return fmt.Errorf("dashboard_aggregation.retention.usage_logs_days must be non-negative")
		}
		if c.DashboardAgg.Retention.HourlyDays < 0 {
			return fmt.Errorf("dashboard_aggregation.retention.hourly_days must be non-negative")
		}
		if c.DashboardAgg.Retention.DailyDays < 0 {
			return fmt.Errorf("dashboard_aggregation.retention.daily_days must be non-negative")
		}
		if c.DashboardAgg.RecomputeDays < 0 {
			return fmt.Errorf("dashboard_aggregation.recompute_days must be non-negative")
		}
	}
	if c.Ops.Cleanup.Enabled && strings.TrimSpace(c.Ops.Cleanup.Schedule) == "" {
		return fmt.Errorf("ops.cleanup.schedule is required")
	}
	if c.Ops.MetricsCollectorCache.TTL < 0 {
		return fmt.Errorf("ops.metrics_collector_cache.ttl must be non-negative")
	}
	if c.Ops.Cleanup.ErrorLogRetentionDays < 0 {
		return fmt.Errorf("ops.cleanup.error_log_retention_days must be non-negative")
	}
	if c.Ops.Cleanup.MinuteMetricsRetentionDays < 0 {
		return fmt.Errorf("ops.cleanup.minute_metrics_retention_days must be non-negative")
	}
	if c.Ops.Cleanup.HourlyMetricsRetentionDays < 0 {
		return fmt.Errorf("ops.cleanup.hourly_metrics_retention_days must be non-negative")
	}
	if c.UsageCleanup.Enabled {
		if c.UsageCleanup.MaxRangeDays <= 0 {
			return fmt.Errorf("usage_cleanup.max_range_days must be positive")
		}
		if c.UsageCleanup.BatchSize <= 0 {
			return fmt.Errorf("usage_cleanup.batch_size must be positive")
		}
		if c.UsageCleanup.WorkerIntervalSeconds <= 0 {
			return fmt.Errorf("usage_cleanup.worker_interval_seconds must be positive")
		}
		if c.UsageCleanup.TaskTimeoutSeconds <= 0 {
			return fmt.Errorf("usage_cleanup.task_timeout_seconds must be positive")
		}
	} else {
		if c.UsageCleanup.MaxRangeDays < 0 {
			return fmt.Errorf("usage_cleanup.max_range_days must be non-negative")
		}
		if c.UsageCleanup.BatchSize < 0 {
			return fmt.Errorf("usage_cleanup.batch_size must be non-negative")
		}
		if c.UsageCleanup.WorkerIntervalSeconds < 0 {
			return fmt.Errorf("usage_cleanup.worker_interval_seconds must be non-negative")
		}
		if c.UsageCleanup.TaskTimeoutSeconds < 0 {
			return fmt.Errorf("usage_cleanup.task_timeout_seconds must be non-negative")
		}
	}
	if c.Gateway.MaxBodySize <= 0 {
		return fmt.Errorf("gateway.max_body_size must be positive")
	}
	if strings.TrimSpace(c.Gateway.ConnectionPoolIsolation) != "" {
		switch c.Gateway.ConnectionPoolIsolation {
		case ConnectionPoolIsolationProxy, ConnectionPoolIsolationAccount, ConnectionPoolIsolationAccountProxy:
		default:
			return fmt.Errorf("gateway.connection_pool_isolation must be one of: %s/%s/%s",
				ConnectionPoolIsolationProxy, ConnectionPoolIsolationAccount, ConnectionPoolIsolationAccountProxy)
		}
	}
	if c.Gateway.MaxIdleConns <= 0 {
		return fmt.Errorf("gateway.max_idle_conns must be positive")
	}
	if c.Gateway.MaxIdleConnsPerHost <= 0 {
		return fmt.Errorf("gateway.max_idle_conns_per_host must be positive")
	}
	if c.Gateway.MaxConnsPerHost < 0 {
		return fmt.Errorf("gateway.max_conns_per_host must be non-negative")
	}
	if c.Gateway.IdleConnTimeoutSeconds <= 0 {
		return fmt.Errorf("gateway.idle_conn_timeout_seconds must be positive")
	}
	if c.Gateway.IdleConnTimeoutSeconds > 180 {
		log.Printf("Warning: gateway.idle_conn_timeout_seconds is %d (> 180). Consider 60-120 seconds for better connection reuse.", c.Gateway.IdleConnTimeoutSeconds)
	}
	if c.Gateway.MaxUpstreamClients <= 0 {
		return fmt.Errorf("gateway.max_upstream_clients must be positive")
	}
	if c.Gateway.ClientIdleTTLSeconds <= 0 {
		return fmt.Errorf("gateway.client_idle_ttl_seconds must be positive")
	}
	if c.Gateway.ConcurrencySlotTTLMinutes <= 0 {
		return fmt.Errorf("gateway.concurrency_slot_ttl_minutes must be positive")
	}
	if c.Gateway.StreamDataIntervalTimeout < 0 {
		return fmt.Errorf("gateway.stream_data_interval_timeout must be non-negative")
	}
	if c.Gateway.StreamDataIntervalTimeout != 0 &&
		(c.Gateway.StreamDataIntervalTimeout < 30 || c.Gateway.StreamDataIntervalTimeout > 300) {
		return fmt.Errorf("gateway.stream_data_interval_timeout must be 0 or between 30-300 seconds")
	}
	if c.Gateway.StreamKeepaliveInterval < 0 {
		return fmt.Errorf("gateway.stream_keepalive_interval must be non-negative")
	}
	if c.Gateway.StreamKeepaliveInterval != 0 &&
		(c.Gateway.StreamKeepaliveInterval < 5 || c.Gateway.StreamKeepaliveInterval > 30) {
		return fmt.Errorf("gateway.stream_keepalive_interval must be 0 or between 5-30 seconds")
	}
	if c.Gateway.MaxLineSize < 0 {
		return fmt.Errorf("gateway.max_line_size must be non-negative")
	}
	if c.Gateway.MaxLineSize != 0 && c.Gateway.MaxLineSize < 1024*1024 {
		return fmt.Errorf("gateway.max_line_size must be at least 1MB")
	}
	if c.Gateway.Scheduling.StickySessionMaxWaiting <= 0 {
		return fmt.Errorf("gateway.scheduling.sticky_session_max_waiting must be positive")
	}
	if c.Gateway.Scheduling.StickySessionWaitTimeout <= 0 {
		return fmt.Errorf("gateway.scheduling.sticky_session_wait_timeout must be positive")
	}
	if c.Gateway.Scheduling.FallbackWaitTimeout <= 0 {
		return fmt.Errorf("gateway.scheduling.fallback_wait_timeout must be positive")
	}
	if c.Gateway.Scheduling.FallbackMaxWaiting <= 0 {
		return fmt.Errorf("gateway.scheduling.fallback_max_waiting must be positive")
	}
	if c.Gateway.Scheduling.SlotCleanupInterval < 0 {
		return fmt.Errorf("gateway.scheduling.slot_cleanup_interval must be non-negative")
	}
	if c.Gateway.Scheduling.OutboxPollIntervalSeconds <= 0 {
		return fmt.Errorf("gateway.scheduling.outbox_poll_interval_seconds must be positive")
	}
	if c.Gateway.Scheduling.OutboxLagRebuildFailures <= 0 {
		return fmt.Errorf("gateway.scheduling.outbox_lag_rebuild_failures must be positive")
	}
	if c.Gateway.Scheduling.OutboxLagRebuildSeconds < c.Gateway.Scheduling.OutboxLagWarnSeconds {
		return fmt.Errorf("gateway.scheduling.outbox_lag_rebuild_seconds must be >= gateway.scheduling.outbox_lag_warn_seconds")
	}
	if c.Concurrency.PingInterval < 5 || c.Concurrency.PingInterval > 30 {
		return fmt.Errorf("concurrency.ping_interval must be between 5-30 seconds")
	}
	return nil
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return values
	}
	normalized := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func parseStringListEnv(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if strings.HasPrefix(raw, "[") {
		var parsed []string
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, err
		}
		return normalizeStringSlice(parsed), nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '\n', ';':
			return true
		default:
			return false
		}
	})
	return normalizeStringSlice(parts), nil
}

func normalizeInt64Slice(values []int64) []int64 {
	if len(values) == 0 {
		return values
	}
	seen := make(map[int64]struct{}, len(values))
	normalized := make([]int64, 0, len(values))
	for _, v := range values {
		if v <= 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		normalized = append(normalized, v)
	}
	return normalized
}

func parseInt64ListEnv(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	if strings.HasPrefix(raw, "[") {
		var parsed []any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, err
		}
		out := make([]int64, 0, len(parsed))
		for _, item := range parsed {
			switch v := item.(type) {
			case float64:
				// JSON numbers decode to float64.
				if v != float64(int64(v)) {
					return nil, fmt.Errorf("non-integer value %v", v)
				}
				out = append(out, int64(v))
			case string:
				trimmed := strings.TrimSpace(v)
				if trimmed == "" {
					continue
				}
				n, err := strconv.ParseInt(trimmed, 10, 64)
				if err != nil {
					return nil, err
				}
				out = append(out, n)
			default:
				return nil, fmt.Errorf("invalid value type %T", item)
			}
		}
		return normalizeInt64Slice(out), nil
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '\n', ';':
			return true
		default:
			return false
		}
	})
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		n, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return normalizeInt64Slice(out), nil
}

func isWeakJWTSecret(secret string) bool {
	lower := strings.ToLower(strings.TrimSpace(secret))
	if lower == "" {
		return true
	}
	weak := map[string]struct{}{
		"change-me-in-production": {},
		"changeme":                {},
		"secret":                  {},
		"password":                {},
		"123456":                  {},
		"12345678":                {},
		"admin":                   {},
		"jwt-secret":              {},
	}
	_, exists := weak[lower]
	return exists
}

func generateJWTSecret(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = 32
	}
	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
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
	v.AddConfigPath("./backend")
	v.AddConfigPath("./backend/config")
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

// ValidateAbsoluteHTTPURL 验证是否为有效的绝对 HTTP(S) URL
func ValidateAbsoluteHTTPURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !u.IsAbs() {
		return fmt.Errorf("must be absolute")
	}
	if !isHTTPScheme(u.Scheme) {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	if strings.TrimSpace(u.Host) == "" {
		return fmt.Errorf("missing host")
	}
	if u.Fragment != "" {
		return fmt.Errorf("must not include fragment")
	}
	return nil
}

// ValidateFrontendRedirectURL 验证前端重定向 URL（可以是绝对 URL 或相对路径）
func ValidateFrontendRedirectURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty url")
	}
	if strings.ContainsAny(raw, "\r\n") {
		return fmt.Errorf("contains invalid characters")
	}
	if strings.HasPrefix(raw, "/") {
		if strings.HasPrefix(raw, "//") {
			return fmt.Errorf("must not start with //")
		}
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !u.IsAbs() {
		return fmt.Errorf("must be absolute http(s) url or relative path")
	}
	if !isHTTPScheme(u.Scheme) {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	if strings.TrimSpace(u.Host) == "" {
		return fmt.Errorf("missing host")
	}
	if u.Fragment != "" {
		return fmt.Errorf("must not include fragment")
	}
	return nil
}

// isHTTPScheme 检查是否为 HTTP 或 HTTPS 协议
func isHTTPScheme(scheme string) bool {
	return strings.EqualFold(scheme, "http") || strings.EqualFold(scheme, "https")
}

func warnIfInsecureURL(field, raw string) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return
	}
	if strings.EqualFold(u.Scheme, "http") {
		log.Printf("Warning: %s uses http scheme; use https in production to avoid token leakage.", field)
	}
}
