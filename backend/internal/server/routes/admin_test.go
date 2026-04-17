package routes

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newAdminRoutesTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	adminHandlers := &handler.AdminHandlers{}
	fillNilPointerFields(adminHandlers)

	router := gin.New()
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	RegisterAdminRoutes(
		v1,
		&handler.Handlers{
			Admin: adminHandlers,
		},
		middleware.AdminAuthMiddleware(func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUserRole), service.RoleAdmin)
			c.Next()
		}),
	)

	return router
}

func fillNilPointerFields(target any) {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return
	}

	elem := value.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Kind() == reflect.Pointer && field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}
}

func TestAdminRoutesCheckMixedChannelPathIsRegistered(t *testing.T) {
	router := newAdminRoutesTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/check-mixed-channel", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.NotEqual(t, http.StatusNotFound, w.Code)
}
