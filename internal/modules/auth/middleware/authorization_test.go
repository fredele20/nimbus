package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authmiddleware "github.com/fredele20/nimbus/internal/modules/auth/middleware"
	"github.com/fredele20/nimbus/internal/platform/testdb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createTestUser(t *testing.T, db *gorm.DB, role, status string) uuid.UUID {
	t.Helper()

	id := uuid.New()

	err := db.Exec(`
		INSERT INTO users (
			id, first_name, last_name,
			email, password, role, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		id,
		"Test",
		"User",
		id.String()+"@example.com",
		"hashed-password",
		role,
		status,
	).Error

	require.NoError(t, err)
	return id
}

func testRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-Test-User-ID"))
		c.Next()
	})

	router.GET(
		"/wallets",
		authmiddleware.LoadUser(db),
		authmiddleware.RequireActiveUser(),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	router.POST(
		"/admin/wallets/freeze",
		authmiddleware.LoadUser(db),
		authmiddleware.RequireRole("admin"),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	return router
}

func performRequest(router *gin.Engine, method, path string, userID uuid.UUID) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)

	if userID != uuid.Nil {
		request.Header.Set("X-Test-User-ID", userID.String())
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestAuthorizationMiddleware(t *testing.T) {
	db := testdb.NewPostgres(t)

	customerID := createTestUser(
		t, db, "customer", "active",
	)
	adminID := createTestUser(
		t, db, "admin", "active",
	)
	suspendedID := createTestUser(
		t, db, "customer", "suspended",
	)

	router := testRouter(db)

	t.Run("unauthenticated user", func(t *testing.T) {
		response := performRequest(
			router, http.MethodGet, "/wallets", uuid.Nil,
		)
		require.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("active customer", func(t *testing.T) {
		response := performRequest(
			router, http.MethodGet, "/wallets", customerID,
		)
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("suspended customer", func(t *testing.T) {
		response := performRequest(
			router, http.MethodGet, "/wallets", suspendedID,
		)
		require.Equal(t, http.StatusForbidden, response.Code)
	})

	t.Run("customer denied admin access", func(t *testing.T) {
		response := performRequest(
			router,
			http.MethodPost,
			"/admin/wallets/freeze",
			customerID,
		)
		require.Equal(t, http.StatusForbidden, response.Code)
	})

	t.Run("admin permitted", func(t *testing.T) {
		response := performRequest(
			router,
			http.MethodPost,
			"/admin/wallets/freeze",
			adminID,
		)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("revoked admin denied", func(t *testing.T) {
		err := db.Exec(
			"UPDATE users SET role = ? WHERE id = ?",
			"customer",
			adminID,
		).Error
		require.NoError(t, err)

		response := performRequest(
			router,
			http.MethodPost,
			"/admin/wallets/freeze",
			adminID,
		)

		require.Equal(t, http.StatusForbidden, response.Code)
	})
}
