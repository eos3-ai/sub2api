// Package middleware provides HTTP middleware for authentication, authorization, and request processing.
package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewAdminAuthMiddleware 创建管理员认证中间件
func NewAdminAuthMiddleware(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	auditService *service.AuditLogService,
) AdminAuthMiddleware {
	return AdminAuthMiddleware(adminAuth(authService, userService, settingService, auditService))
}

// adminAuth 管理员认证中间件实现
// 支持三种认证方式：
// 1. Admin API Key（读写）: x-api-key: <admin-api-key>
// 2. Admin API Key（只读）: x-api-key: <admin-api-key-read-only>（仅 GET 白名单）
// 3. JWT Token: Authorization: Bearer <jwt-token> (需要后台角色：admin/sales)
func adminAuth(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	auditService *service.AuditLogService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket upgrade requests cannot set Authorization headers in browsers.
		// For admin WebSocket endpoints (e.g. Ops realtime), allow passing the JWT via
		// Sec-WebSocket-Protocol (subprotocol list) using a prefixed token item:
		//   Sec-WebSocket-Protocol: sub2api-admin, jwt.<token>
		if isWebSocketUpgradeRequest(c) {
			if token := extractJWTFromWebSocketSubprotocol(c); token != "" {
				if !validateJWTForAdmin(c, token, authService, userService, settingService, auditService) {
					return
				}
				c.Next()
				return
			}
		}

		// 检查 x-api-key header（Admin API Key 认证）
		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			if !validateAdminAPIKey(c, apiKey, settingService, userService) {
				return
			}
			c.Next()
			return
		}

		// 检查 Authorization header（JWT 认证）
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token := strings.TrimSpace(parts[1])
				if token == "" {
					AbortWithError(c, 401, "UNAUTHORIZED", "Authorization required")
					return
				}
				if !validateJWTForAdmin(c, token, authService, userService, settingService, auditService) {
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

type adminAPIKeyAccess int

const (
	adminAPIKeyAccessReadWrite adminAPIKeyAccess = iota
	adminAPIKeyAccessReadOnly
)

const readOnlyAdminAPIKeyAllowedPathsEnv = "SECURITY_ADMIN_API_KEY_READ_ONLY_ALLOWED_PATHS"

func isReadOnlyAdminAPIKeyMethod(method string) bool {
	return method == http.MethodGet
}

func normalizeReadOnlyAdminAPIKeyPath(raw string) string {
	path := strings.TrimSpace(raw)
	if path != "" && path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func parseReadOnlyAdminAPIKeyAllowedPaths(raw string) map[string]struct{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var values []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return nil
		}
	} else {
		values = strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == '\n' || r == '\r' || r == ';'
		})
	}

	allowed := make(map[string]struct{}, len(values))
	for _, value := range values {
		path := normalizeReadOnlyAdminAPIKeyPath(value)
		if path == "" || !strings.HasPrefix(path, "/") {
			continue
		}
		allowed[path] = struct{}{}
	}
	return allowed
}

func isReadOnlyAdminAPIKeyAllowedPath(rawPath string) bool {
	path := normalizeReadOnlyAdminAPIKeyPath(rawPath)
	allowed := parseReadOnlyAdminAPIKeyAllowedPaths(os.Getenv(readOnlyAdminAPIKeyAllowedPathsEnv))
	_, ok := allowed[path]
	return ok
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

// validateAdminAPIKey 验证管理员 API Key
func validateAdminAPIKey(
	c *gin.Context,
	key string,
	settingService *service.SettingService,
	userService *service.UserService,
) bool {
	access, matched, err := matchAdminAPIKeyAccess(c.Request.Context(), key, settingService)
	if err != nil {
		AbortWithError(c, 500, "INTERNAL_ERROR", "Internal server error")
		return false
	}

	// 未配置或不匹配，统一返回相同错误（避免信息泄露）
	if !matched {
		AbortWithError(c, 401, "INVALID_ADMIN_KEY", "Invalid admin API key")
		return false
	}

	if access == adminAPIKeyAccessReadOnly {
		if !isReadOnlyAdminAPIKeyMethod(c.Request.Method) || !isReadOnlyAdminAPIKeyAllowedPath(c.Request.URL.Path) {
			AbortWithError(c, 403, "ADMIN_API_KEY_READ_ONLY", "Admin API key is read-only")
			return false
		}
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
	c.Set(ContextKeyAuthEmail, admin.Email)
	c.Set("auth_method", "admin_api_key")
	return true
}

// validateJWTForAdmin 验证 JWT 并检查后台角色权限（admin/sales）
func validateJWTForAdmin(
	c *gin.Context,
	token string,
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	auditService *service.AuditLogService,
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

	// 校验 TokenVersion，确保管理员改密后旧 token 失效
	if claims.TokenVersion != user.TokenVersion {
		AbortWithError(c, 401, "TOKEN_REVOKED", "Token has been revoked (password changed)")
		return false
	}

	// 会话绑定校验：IP/UA 任一变化即撤销会话（功能可在系统设置中关闭）
	if !enforceSessionBinding(c, authService, settingService, auditService, claims) {
		return false
	}

	// 检查后台权限
	if !IsBackofficeRole(user.Role) {
		AbortWithError(c, 403, "FORBIDDEN", "Backoffice access required")
		return false
	}

	c.Set(string(ContextKeyUser), AuthSubject{
		UserID:      user.ID,
		Concurrency: user.Concurrency,
	})
	c.Set(string(ContextKeyUserRole), user.Role)
	c.Set(ContextKeyAuthEmail, user.Email)
	c.Set(ContextKeySessionID, claims.SessionID)
	c.Set("auth_method", "jwt")

	return true
}
