// Package middleware provides HTTP middleware for authentication, authorization, and request processing.
package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewAdminAuthMiddleware 创建管理员认证中间件
func NewAdminAuthMiddleware(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	cfg *config.Config,
) AdminAuthMiddleware {
	return AdminAuthMiddleware(adminAuth(authService, userService, settingService, cfg))
}

// adminAuth 管理员认证中间件实现
// 支持三种认证方式：
// 1. Admin API Key（读写）：x-api-key: <admin-api-key>
// 2. Admin API Key（只读）：x-api-key: <admin-api-key-read-only>（仅允许 GET）
// 3. JWT Token：Authorization: Bearer <jwt-token>（需要管理员角色）
func adminAuth(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	cfg *config.Config,
) gin.HandlerFunc {
	readOnlyAllowlist := newReadOnlyAdminAPIKeyAllowlist(cfg)

	return func(c *gin.Context) {
		// WebSocket upgrade requests cannot set Authorization headers in browsers.
		// For admin WebSocket endpoints (e.g. Ops realtime), allow passing the JWT via
		// Sec-WebSocket-Protocol (subprotocol list) using a prefixed token item:
		//   Sec-WebSocket-Protocol: sub2api-admin, jwt.<token>
		if isWebSocketUpgradeRequest(c) {
			if token := extractJWTFromWebSocketSubprotocol(c); token != "" {
				if !validateJWTForAdmin(c, token, authService, userService) {
					return
				}
				c.Next()
				return
			}
		}

		// 检查 x-api-key header（Admin API Key 认证）
		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			access, matched, err := matchAdminAPIKeyAccess(c.Request.Context(), apiKey, settingService)
			if err != nil {
				AbortWithError(c, 500, "INTERNAL_ERROR", "Internal server error")
				return
			}
			if !matched {
				AbortWithError(c, 401, "INVALID_ADMIN_KEY", "Invalid admin API key")
				return
			}

			if access == adminAPIKeyAccessReadOnly {
				// Read-only key: allow only GET and a configurable allowlist of endpoints.
				if !isReadOnlyAdminAPIKeyMethod(c.Request.Method) || !readOnlyAllowlist.Allows(c.Request.URL.Path) {
					AbortWithError(c, 403, "ADMIN_API_KEY_READ_ONLY", "Admin API key is read-only")
					return
				}
			}

			if !setAdminAPIKeyContext(c, userService) {
				return
			}
			c.Next()
			return
		}

		// 检查 Authorization header（JWT 认证）
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				if !validateJWTForAdmin(c, parts[1], authService, userService) {
					return
				}
				c.Next()
				return
			}
		}

		// 无有效认证信息
		AbortWithError(c, 401, "UNAUTHORIZED", "Authorization required")
	}
}

type adminAPIKeyAccess int

const (
	adminAPIKeyAccessReadWrite adminAPIKeyAccess = iota
	adminAPIKeyAccessReadOnly
)

func isReadOnlyAdminAPIKeyMethod(method string) bool {
	switch method {
	case http.MethodGet:
		return true
	default:
		return false
	}
}

type readOnlyAdminAPIKeyAllowlist struct {
	exact    map[string]struct{}
	prefixes []string
}

func newReadOnlyAdminAPIKeyAllowlist(cfg *config.Config) readOnlyAdminAPIKeyAllowlist {
	paths := []string{
		"/api/v1/admin/users/export",
		"/api/v1/admin/usage",
		"/api/v1/admin/payment/orders/export",
	}
	prefixes := []string{}
	if cfg != nil {
		paths = cfg.Security.AdminAPIKeyReadOnly.AllowedPaths
		prefixes = cfg.Security.AdminAPIKeyReadOnly.AllowedPathPrefixes
	}

	exact := make(map[string]struct{}, len(paths))
	for _, raw := range paths {
		p := normalizePathForAllowlist(raw)
		if p == "" {
			continue
		}
		exact[p] = struct{}{}
	}
	normalizedPrefixes := make([]string, 0, len(prefixes))
	for _, raw := range prefixes {
		p := normalizePathForAllowlist(raw)
		if p == "" {
			continue
		}
		normalizedPrefixes = append(normalizedPrefixes, p)
	}

	return readOnlyAdminAPIKeyAllowlist{
		exact:    exact,
		prefixes: normalizedPrefixes,
	}
}

func (a readOnlyAdminAPIKeyAllowlist) Allows(rawPath string) bool {
	path := normalizePathForAllowlist(rawPath)
	if path == "" {
		return false
	}
	if _, ok := a.exact[path]; ok {
		return true
	}
	for _, prefix := range a.prefixes {
		if prefix == "" {
			continue
		}
		if prefix == "/" {
			return true
		}
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		// Segment boundary match: "/a/b" matches "/a/b" and "/a/b/c", but not "/a/bc".
		if len(path) == len(prefix) {
			return true
		}
		if len(path) > len(prefix) && path[len(prefix)] == '/' {
			return true
		}
	}
	return false
}

func normalizePathForAllowlist(raw string) string {
	// Normalize trailing slashes to avoid accidental denial on "/path/".
	path := strings.TrimSpace(raw)
	if path != "" && path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func isWebSocketUpgradeRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	// RFC6455 handshake uses:
	//   Connection: Upgrade
	//   Upgrade: websocket
	upgrade := strings.ToLower(strings.TrimSpace(c.GetHeader("Upgrade")))
	if upgrade != "websocket" {
		return false
	}
	connection := strings.ToLower(c.GetHeader("Connection"))
	return strings.Contains(connection, "upgrade")
}

func extractJWTFromWebSocketSubprotocol(c *gin.Context) string {
	if c == nil {
		return ""
	}
	raw := strings.TrimSpace(c.GetHeader("Sec-WebSocket-Protocol"))
	if raw == "" {
		return ""
	}

	// The header is a comma-separated list of tokens. We reserve the prefix "jwt."
	// for carrying the admin JWT.
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if strings.HasPrefix(p, "jwt.") {
			token := strings.TrimSpace(strings.TrimPrefix(p, "jwt."))
			if token != "" {
				return token
			}
		}
	}
	return ""
}

func matchAdminAPIKeyAccess(ctx context.Context, key string, settingService *service.SettingService) (adminAPIKeyAccess, bool, error) {
	if settingService == nil {
		return adminAPIKeyAccessReadWrite, false, nil
	}

	storedKeyReadWrite, err := settingService.GetAdminAPIKey(ctx)
	if err != nil {
		return adminAPIKeyAccessReadWrite, false, err
	}
	storedKeyReadOnly, err := settingService.GetAdminAPIKeyReadOnly(ctx)
	if err != nil {
		return adminAPIKeyAccessReadWrite, false, err
	}

	if storedKeyReadWrite != "" && subtle.ConstantTimeCompare([]byte(key), []byte(storedKeyReadWrite)) == 1 {
		return adminAPIKeyAccessReadWrite, true, nil
	}
	if storedKeyReadOnly != "" && subtle.ConstantTimeCompare([]byte(key), []byte(storedKeyReadOnly)) == 1 {
		return adminAPIKeyAccessReadOnly, true, nil
	}
	return adminAPIKeyAccessReadWrite, false, nil
}

func setAdminAPIKeyContext(c *gin.Context, userService *service.UserService) bool {
	if c == nil || c.Request == nil {
		return false
	}

	// 获取真实的管理员用户
	admin, err := userService.GetFirstAdmin(c.Request.Context())
	if err != nil {
		AbortWithError(c, 500, "INTERNAL_ERROR", "No admin user found")
		return false
	}

	c.Set(string(ContextKeyUser), AuthSubject{
		UserID:      admin.ID,
		Concurrency: admin.Concurrency,
	})
	c.Set(string(ContextKeyUserRole), admin.Role)
	c.Set("auth_method", "admin_api_key")
	return true
}

// validateJWTForAdmin 验证 JWT 并检查管理员权限
func validateJWTForAdmin(
	c *gin.Context,
	token string,
	authService *service.AuthService,
	userService *service.UserService,
) bool {
	// 验证 JWT token
	claims, err := authService.ValidateToken(token)
	if err != nil {
		if errors.Is(err, service.ErrTokenExpired) {
			AbortWithError(c, 401, "TOKEN_EXPIRED", "Token has expired")
			return false
		}
		AbortWithError(c, 401, "INVALID_TOKEN", "Invalid token")
		return false
	}

	// 从数据库获取用户
	user, err := userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		AbortWithError(c, 401, "USER_NOT_FOUND", "User not found")
		return false
	}

	// 检查用户状态
	if !user.IsActive() {
		AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
		return false
	}

	// 检查管理员权限
	if !user.IsAdmin() {
		AbortWithError(c, 403, "FORBIDDEN", "Admin access required")
		return false
	}

	c.Set(string(ContextKeyUser), AuthSubject{
		UserID:      user.ID,
		Concurrency: user.Concurrency,
	})
	c.Set(string(ContextKeyUserRole), user.Role)
	c.Set("auth_method", "jwt")

	return true
}
