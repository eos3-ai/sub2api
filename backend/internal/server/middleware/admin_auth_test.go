package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminAPIKeyAccessModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminReadWriteKey := "admin-test-key-rw"
	adminReadOnlyKey := "admin-test-key-ro"

	settingService := service.NewSettingService(&stubSettingRepo{
		values: map[string]string{
			service.SettingKeyAdminAPIKey:         adminReadWriteKey,
			service.SettingKeyAdminAPIKeyReadOnly: adminReadOnlyKey,
		},
	}, nil)

	userService := service.NewUserService(&stubUserRepo{
		firstAdmin: &service.User{
			ID:          1,
			Role:        service.RoleAdmin,
			Status:      service.StatusActive,
			Concurrency: 1,
		},
	}, nil)

	router := gin.New()
	router.Use(adminAuth(nil, userService, settingService))

	// Allowed paths for the read-only admin key.
	router.GET("/api/v1/admin/users/export", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/admin/usage", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/admin/payment/orders/export", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Not allowed for the read-only admin key.
	router.GET("/api/v1/admin/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Method restriction: read-only key should reject non-GET even on allowed paths.
	router.POST("/api/v1/admin/users/export", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	t.Run("read_write_key_allows_get", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/export", nil)
		req.Header.Set("x-api-key", adminReadWriteKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("read_write_key_allows_get_on_non_allowlisted_path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		req.Header.Set("x-api-key", adminReadWriteKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("read_write_key_allows_post", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/export", nil)
		req.Header.Set("x-api-key", adminReadWriteKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("read_only_key_allows_get_on_allowlisted_path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/export", nil)
		req.Header.Set("x-api-key", adminReadOnlyKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("read_only_key_rejects_get_on_non_allowlisted_path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		req.Header.Set("x-api-key", adminReadOnlyKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "ADMIN_API_KEY_READ_ONLY")
	})

	t.Run("read_only_key_rejects_post", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/export", nil)
		req.Header.Set("x-api-key", adminReadOnlyKey)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "ADMIN_API_KEY_READ_ONLY")
	})

	t.Run("rejects_invalid_key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/export", nil)
		req.Header.Set("x-api-key", "admin-wrong-key")
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "INVALID_ADMIN_KEY")
	})
}

type stubSettingRepo struct {
	values map[string]string
}

func (r *stubSettingRepo) Get(ctx context.Context, key string) (*service.Setting, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	return &service.Setting{Key: key, Value: value}, nil
}

func (r *stubSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	if r.values == nil {
		return "", service.ErrSettingNotFound
	}
	value, ok := r.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (r *stubSettingRepo) Set(ctx context.Context, key, value string) error {
	return errors.New("not implemented")
}

func (r *stubSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	return nil, errors.New("not implemented")
}

func (r *stubSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	return errors.New("not implemented")
}

func (r *stubSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	return nil, errors.New("not implemented")
}

func (r *stubSettingRepo) Delete(ctx context.Context, key string) error {
	return errors.New("not implemented")
}

type stubUserRepo struct {
	firstAdmin *service.User
}

func (r *stubUserRepo) Create(ctx context.Context, user *service.User) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) GetByID(ctx context.Context, id int64) (*service.User, error) {
	return nil, errors.New("not implemented")
}

func (r *stubUserRepo) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	return nil, errors.New("not implemented")
}

func (r *stubUserRepo) GetByUsername(ctx context.Context, username string) (*service.User, error) {
	return nil, errors.New("not implemented")
}

func (r *stubUserRepo) GetEmailsByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	return nil, errors.New("not implemented")
}

func (r *stubUserRepo) GetFirstAdmin(ctx context.Context) (*service.User, error) {
	if r.firstAdmin == nil {
		return nil, service.ErrUserNotFound
	}
	clone := *r.firstAdmin
	return &clone, nil
}

func (r *stubUserRepo) Update(ctx context.Context, user *service.User) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) List(ctx context.Context, params pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *stubUserRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *stubUserRepo) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) DeductBalance(ctx context.Context, id int64, amount float64) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	return errors.New("not implemented")
}

func (r *stubUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *stubUserRepo) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	return 0, errors.New("not implemented")
}
