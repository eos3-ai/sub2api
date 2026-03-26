package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupUserHandlerRouterWithRole(role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if role != "" {
		router.Use(func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUserRole), role)
			c.Next()
		})
	}

	adminSvc := newStubAdminService()
	userHandler := NewUserHandler(adminSvc, nil)
	router.POST("/api/v1/admin/users/:id/balance", userHandler.UpdateBalance)
	router.GET("/api/v1/admin/users/:id/api-keys", userHandler.GetUserAPIKeys)
	return router
}

func TestSalesUpdateBalanceConstraints(t *testing.T) {
	router := setupUserHandlerRouterWithRole("sales")

	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "allow_small_add_with_notes",
			body:       `{"balance":8,"operation":"add","notes":"manual compensation"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "reject_amount_over_limit",
			body:       `{"balance":8.01,"operation":"add","notes":"too much"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "reject_missing_notes",
			body:       `{"balance":2,"operation":"add","notes":"   "}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reject_non_add_operation",
			body:       `{"balance":2,"operation":"set","notes":"not allowed"}`,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/1/balance", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestAdminUpdateBalanceUnaffected(t *testing.T) {
	router := setupUserHandlerRouterWithRole("admin")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/1/balance", bytes.NewBufferString(`{"balance":100,"operation":"set"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUserAPIKeysMasksForSales(t *testing.T) {
	router := setupUserHandlerRouterWithRole("sales")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/1/api-keys", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Items []struct {
				Key string `json:"key"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, "****", resp.Data.Items[0].Key)
}

func TestGetUserAPIKeysAdminKeepsRawKey(t *testing.T) {
	router := setupUserHandlerRouterWithRole("admin")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/1/api-keys", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Items []struct {
				Key string `json:"key"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, "sk-test", resp.Data.Items[0].Key)
}
