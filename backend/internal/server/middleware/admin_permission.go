package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type adminPermission string

const (
	adminPermissionDashboardRead       adminPermission = "dashboard:read"
	adminPermissionUserRead            adminPermission = "user:read"
	adminPermissionUserExport          adminPermission = "user:export"
	adminPermissionPaymentRead         adminPermission = "payment:read"
	adminPermissionPaymentExport       adminPermission = "payment:export"
	adminPermissionInvoiceRead         adminPermission = "invoice:read"
	adminPermissionInvoiceStatusUpdate adminPermission = "invoice:status_update"
	adminPermissionUsageRead           adminPermission = "usage:read"
	adminPermissionUsageExport         adminPermission = "usage:export"
)

var salesAllowedAdminPermissions = map[adminPermission]struct{}{
	adminPermissionDashboardRead:       {},
	adminPermissionUserRead:            {},
	adminPermissionUserExport:          {},
	adminPermissionPaymentRead:         {},
	adminPermissionPaymentExport:       {},
	adminPermissionInvoiceRead:         {},
	adminPermissionInvoiceStatusUpdate: {},
	adminPermissionUsageRead:           {},
	adminPermissionUsageExport:         {},
}

// IsBackofficeRole reports whether a role can enter the admin route group.
func IsBackofficeRole(role string) bool {
	switch strings.TrimSpace(role) {
	case service.RoleAdmin, service.RoleSales:
		return true
	default:
		return false
	}
}

// RequireAdminPermission enforces route-level permissions for non-admin
// backoffice roles. Admin has full access; sales is restricted by a static
// allowlist that matches the migration plan.
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

		if role != service.RoleSales {
			AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Backoffice access required")
			return
		}

		perm, matched := resolveAdminPermission(c.Request.Method, c.FullPath())
		if !matched {
			AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Sales permission denied")
			return
		}
		if _, allowed := salesAllowedAdminPermissions[perm]; !allowed {
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

	if hasPathPrefix(path, "/admin/dashboard") {
		if method == http.MethodGet {
			return adminPermissionDashboardRead, true
		}
		if method == http.MethodPost && (path == "/admin/dashboard/users-usage" || path == "/admin/dashboard/api-keys-usage") {
			return adminPermissionDashboardRead, true
		}
		return "", false
	}

	if hasPathPrefix(path, "/admin/users") {
		if method == http.MethodGet {
			if path == "/admin/users/export" {
				return adminPermissionUserExport, true
			}
			switch path {
			case "/admin/users",
				"/admin/users/:id",
				"/admin/users/:id/api-keys",
				"/admin/users/:id/usage",
				"/admin/users/:id/balance-history",
				"/admin/users/:id/rpm-status",
				"/admin/users/:id/subscriptions",
				"/admin/users/:id/attributes":
				return adminPermissionUserRead, true
			}
		}
		return "", false
	}

	if hasPathPrefix(path, "/admin/user-attributes") {
		if method == http.MethodGet && path == "/admin/user-attributes" {
			return adminPermissionUserRead, true
		}
		if method == http.MethodPost && path == "/admin/user-attributes/batch" {
			return adminPermissionUserRead, true
		}
		return "", false
	}

	if hasPathPrefix(path, "/admin/payment/orders") {
		if method == http.MethodGet {
			if path == "/admin/payment/orders/export" {
				return adminPermissionPaymentExport, true
			}
			return adminPermissionPaymentRead, true
		}
		return "", false
	}

	if hasPathPrefix(path, "/admin/invoices") {
		if method == http.MethodGet {
			return adminPermissionInvoiceRead, true
		}
		if method == http.MethodPost && isInvoiceStatusMutationPath(path) {
			return adminPermissionInvoiceStatusUpdate, true
		}
		return "", false
	}

	if hasPathPrefix(path, "/admin/usage") {
		if method == http.MethodGet {
			if path == "/admin/usage/export" {
				return adminPermissionUsageExport, true
			}
			return adminPermissionUsageRead, true
		}
		return "", false
	}

	return "", false
}

func isInvoiceStatusMutationPath(path string) bool {
	switch path {
	case "/admin/invoices/:id/approve",
		"/admin/invoices/:id/reject",
		"/admin/invoices/:id/issue":
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
