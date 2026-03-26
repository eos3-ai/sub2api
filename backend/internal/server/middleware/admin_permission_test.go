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
	require.True(t, IsBackofficeRole("operator"))
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

func TestRequireAdminPermission_OperatorAllowDenyMatrix(t *testing.T) {
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
			name:       "allow_dashboard_batch_post",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/dashboard/users-usage",
			requestURL: "/api/v1/admin/dashboard/users-usage",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_user_read",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/users/:id/api-keys",
			requestURL: "/api/v1/admin/users/9/api-keys",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_small_compensation",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/users/:id/balance",
			requestURL: "/api/v1/admin/users/9/balance",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_user_attribute_definitions_read",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/user-attributes",
			requestURL: "/api/v1/admin/user-attributes",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_user_attribute_batch_read",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/user-attributes/batch",
			requestURL: "/api/v1/admin/user-attributes/batch",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_ops_dashboard",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/ops/dashboard/snapshot-v2",
			requestURL: "/api/v1/admin/ops/dashboard/snapshot-v2",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow_ops_alert_events_read",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/ops/alert-events/:id",
			requestURL: "/api/v1/admin/ops/alert-events/12",
			wantStatus: http.StatusOK,
		},
		{
			name:       "deny_ops_request_drilldown",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/ops/requests",
			requestURL: "/api/v1/admin/ops/requests",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "deny_ops_write",
			method:     http.MethodPost,
			routePath:  "/api/v1/admin/ops/alert-rules",
			requestURL: "/api/v1/admin/ops/alert-rules",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "deny_settings",
			method:     http.MethodGet,
			routePath:  "/api/v1/admin/settings",
			requestURL: "/api/v1/admin/settings",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "deny_delete_user",
			method:     http.MethodDelete,
			routePath:  "/api/v1/admin/users/:id",
			requestURL: "/api/v1/admin/users/9",
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
				c.Set(string(ContextKeyUserRole), "operator")
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
