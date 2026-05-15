//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIsBackofficeRole(t *testing.T) {
	require.True(t, IsBackofficeRole("admin"))
	require.True(t, IsBackofficeRole("sales"))
	require.False(t, IsBackofficeRole("operator"))
	require.False(t, IsBackofficeRole("user"))
	require.False(t, IsBackofficeRole(""))
}

func TestRequireAdminPermission_AdminHasFullAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), "admin")
		c.Next()
	})
	r.Use(RequireAdminPermission())
	r.GET("/api/v1/admin/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdminPermission_SalesAllowDenyMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		routePath  string
		requestURL string
		wantStatus int
	}{
		{
			name:       "allow_dashboard_get",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/dashboard/stats",
			requestURL: "/api/v1/admin/dashboard/stats",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_user_read",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/users",
			requestURL: "/api/v1/admin/users",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_user_export",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/users/export",
			requestURL: "/api/v1/admin/users/export",
			wantStatus: http.StatusOK,
		},
		{
			name:       "deny_create_user",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/users",
			requestURL: "/api/v1/admin/users",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "allow_payment_read",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/payment/orders",
			requestURL: "/api/v1/admin/payment/orders",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_payment_export",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/payment/orders/export",
			requestURL: "/api/v1/admin/payment/orders/export",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_usage_export",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/usage/export",
			requestURL: "/api/v1/admin/usage/export",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_invoice_approve",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/invoices/:id/approve",
			requestURL: "/api/v1/admin/invoices/9/approve",
			wantStatus: http.StatusOK,
		},
		{
			name:       "deny_settings",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/settings",
			requestURL: "/api/v1/admin/settings",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "allow_without_api_prefix",
			method:     http.MethodGet,
			routePath:  "/admin/dashboard/stats",
			requestURL: "/admin/dashboard/stats",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set(string(ContextKeyUserRole), "sales")
				c.Next()
			})
			r.Use(RequireAdminPermission())
			r.Handle(tc.method, tc.routePath, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.requestURL, nil)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
