package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type adminPermission string

const (
	adminPermissionDashboardRead         adminPermission = "dashboard:read"
	adminPermissionUserRead              adminPermission = "user:read"
	adminPermissionUserSmallCompensation adminPermission = "user:small_compensation"
	adminPermissionUsageRead             adminPermission = "usage:read"
	adminPermissionOpsRead               adminPermission = "ops:read"
	adminPermissionPaymentRead           adminPermission = "payment:read"
	adminPermissionInvoiceRead           adminPermission = "invoice:read"
	adminPermissionInvoiceStatusUpdate   adminPermission = "invoice:status_update"
	adminPermissionAnnouncementManage    adminPermission = "announcement:manage"
)

var operatorAllowedPermissions = map[adminPermission]struct{}{
	adminPermissionDashboardRead:         {},
	adminPermissionUserRead:              {},
	adminPermissionUserSmallCompensation: {},
	adminPermissionUsageRead:             {},
	adminPermissionOpsRead:               {},
	adminPermissionPaymentRead:           {},
	adminPermissionInvoiceRead:           {},
	adminPermissionInvoiceStatusUpdate:   {},
	adminPermissionAnnouncementManage:    {},
}

// IsBackofficeRole reports whether a role is allowed to access backoffice routes.
func IsBackofficeRole(role string) bool {
	switch strings.TrimSpace(role) {
	case service.RoleAdmin, service.RoleOperator:
		return true
	default:
		return false
	}
}

// RequireAdminPermission enforces fine-grained admin-route permission checks.
// Admin has full access; sales is restricted by static allowlist.
func RequireAdminPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRoleFromContext(c)
		if !ok || strings.TrimSpace(role) == "" {
			AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User role not found")
			return
		}

		if role == service.RoleAdmin {
			c.Next()
			return
		}

		if role != service.RoleOperator {
			AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Backoffice access required")
			return
		}

		perm, matched := resolveAdminPermission(c.Request.Method, c.FullPath())
		if !matched {
			AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Sales permission denied")
			return
		}
		if _, allowed := operatorAllowedPermissions[perm]; !allowed {
			AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Sales permission denied")
			return
		}
		c.Next()
	}
}

func resolveAdminPermission(method, fullPath string) (adminPermission, bool) {
	path := normalizeAdminRoutePath(fullPath)
	if path == "" {
		return "", false
	}

	// Dashboard
	if hasPathPrefix(path, "/admin/dashboard") {
		if method == http.MethodGet {
			return adminPermissionDashboardRead, true
		}
		if method == http.MethodPost && (path == "/admin/dashboard/users-usage" || path == "/admin/dashboard/api-keys-usage") {
			return adminPermissionDashboardRead, true
		}
		return "", false
	}

	// User management (readonly + small compensation)
	if hasPathPrefix(path, "/admin/users") {
		if method == http.MethodGet {
			switch path {
			case "/admin/users",
				"/admin/users/export",
				"/admin/users/:id",
				"/admin/users/:id/api-keys",
				"/admin/users/:id/usage",
				"/admin/users/:id/balance-history",
				"/admin/users/:id/attributes":
				return adminPermissionUserRead, true
			}
		}
		if method == http.MethodPost && path == "/admin/users/:id/balance" {
			return adminPermissionUserSmallCompensation, true
		}
		return "", false
	}

	// User attributes (readonly)
	if hasPathPrefix(path, "/admin/user-attributes") {
		if method == http.MethodGet && path == "/admin/user-attributes" {
			return adminPermissionUserRead, true
		}
		if method == http.MethodPost && path == "/admin/user-attributes/batch" {
			return adminPermissionUserRead, true
		}
		return "", false
	}

	// Usage records
	if hasPathPrefix(path, "/admin/usage") {
		if method == http.MethodGet {
			return adminPermissionUsageRead, true
		}
		return "", false
	}

	// Ops monitoring (readonly safe subset)
	if hasPathPrefix(path, "/admin/ops") {
		if method == http.MethodGet && isOperatorAllowedOpsPath(path) {
			return adminPermissionOpsRead, true
		}
		return "", false
	}

	// Payment orders (readonly)
	if hasPathPrefix(path, "/admin/payment/orders") {
		if method == http.MethodGet {
			return adminPermissionPaymentRead, true
		}
		return "", false
	}

	// Invoices (readonly + status actions)
	if hasPathPrefix(path, "/admin/invoices") {
		if method == http.MethodGet {
			return adminPermissionInvoiceRead, true
		}
		if method == http.MethodPost && isInvoiceStatusMutationPath(path) {
			return adminPermissionInvoiceStatusUpdate, true
		}
		return "", false
	}

	// Announcements (publish / update / delete allowed)
	if hasPathPrefix(path, "/admin/announcements") {
		switch method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
			return adminPermissionAnnouncementManage, true
		default:
			return "", false
		}
	}

	return "", false
}

func isOperatorAllowedOpsPath(path string) bool {
	switch path {
	case "/admin/ops/concurrency",
		"/admin/ops/user-concurrency",
		"/admin/ops/account-availability",
		"/admin/ops/realtime-traffic",
		"/admin/ops/alert-rules",
		"/admin/ops/alert-events",
		"/admin/ops/alert-events/:id",
		"/admin/ops/settings/metric-thresholds",
		"/admin/ops/advanced-settings",
		"/admin/ops/system-logs/health",
		"/admin/ops/ws/qps":
		return true
	}
	if hasPathPrefix(path, "/admin/ops/dashboard") {
		return true
	}
	return false
}

func isInvoiceStatusMutationPath(path string) bool {
	switch path {
	case "/admin/invoices/:id/approve",
		"/admin/invoices/:id/reject",
		"/admin/invoices/:id/issue",
		"/admin/invoices/:id/retry-issue":
		return true
	default:
		return false
	}
}

func normalizeAdminRoutePath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "/api/v1") {
		path = strings.TrimPrefix(path, "/api/v1")
	}
	if path == "" {
		path = "/"
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func hasPathPrefix(path, prefix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	if len(path) == len(prefix) {
		return true
	}
	return path[len(prefix)] == '/'
}
