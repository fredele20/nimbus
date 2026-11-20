package wallet_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fredele20/nimbus/internal/modules/auth/token"
	"github.com/fredele20/nimbus/internal/modules/wallet"
	"github.com/fredele20/nimbus/internal/modules/wallet/controller"
	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/fredele20/nimbus/internal/modules/wallet/service"
	"github.com/fredele20/nimbus/internal/platform/testdb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testJWTSecret = "nimbus-wallet-integration-test-secret"

type walletHTTPTest struct {
	t      *testing.T
	db     *gorm.DB
	router *gin.Engine
	repo   repository.Repository
}

func newWalletHTTPTest(t *testing.T) *walletHTTPTest {
	t.Helper()

	db := testdb.NewPostgres(t)
	repo := repository.New(db)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")

	walletController := controller.New(
		service.New(repo),
		service.NewAdmin(repo),
	)

	wallet.RegisterRoutes(
		v1,
		walletController,
		testJWTSecret,
		db,
	)

	return &walletHTTPTest{
		t:      t,
		db:     db,
		router: router,
		repo:   repo,
	}
}

func (s *walletHTTPTest) createUser(role, status string) uuid.UUID {
	s.t.Helper()

	id := uuid.New()

	err := s.db.Exec(`
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

	require.NoError(s.t, err)
	return id
}

func (s *walletHTTPTest) createWallet(ownerID uuid.UUID) model.Wallet {
	s.t.Helper()

	w := model.Wallet{
		UserID:   ownerID,
		Currency: "NGN",
	}

	require.NoError(
		s.t,
		s.repo.Create(context.Background(), &w),
	)

	return w
}

func (s *walletHTTPTest) accessToken(userID uuid.UUID) string {
	s.t.Helper()

	value, err := token.New(testJWTSecret).GenerateAccessToken(userID.String())

	require.NoError(s.t, err)
	return value
}

func (s *walletHTTPTest) request(method, path string, userID uuid.UUID, body string) *httptest.ResponseRecorder {
	s.t.Helper()

	req := httptest.NewRequest(
		method,
		path,
		bytes.NewBufferString(body),
	)

	if userID != uuid.Nil {
		req.Header.Set("Authorization", "Bearer "+s.accessToken(userID))
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)

	return recorder
}

func TestWalletHTTPAuthentication(t *testing.T) {
	s := newWalletHTTPTest(t)

	t.Run("missing JWT", func(t *testing.T) {
		res := s.request(http.MethodGet, "/api/v1/wallets", uuid.Nil, "")

		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("invalid JWT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets", nil)

		req.Header.Set("Authorization", "Bearer invalid-token")

		recorder := httptest.NewRecorder()
		s.router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("suspended user", func(t *testing.T) {
		userID := s.createUser("customer", "suspended")

		res := s.request(http.MethodGet, "/api/v1/wallets", userID, "")

		require.Equal(t, http.StatusForbidden, res.Code)
	})
}

func TestWalletHTTPOwnership(t *testing.T) {
	s := newWalletHTTPTest(t)

	ownerID := s.createUser("customer", "active")
	otherID := s.createUser("customer", "active")

	w := s.createWallet(ownerID)

	t.Run("owner can retrieve wallet", func(t *testing.T) {
		res := s.request(http.MethodGet, "/api/v1/wallets/"+w.ID.String(), ownerID, "")

		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("another user cannot retrieve wallet", func(t *testing.T) {
		res := s.request(http.MethodGet, "/api/v1/wallets/"+w.ID.String(), otherID, "")

		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("another user cannot freeze wallet", func(t *testing.T) {
		res := s.request(http.MethodPost, "/api/v1/wallets/"+w.ID.String()+"/freeze", otherID, "")

		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("invalid wallet ID", func(t *testing.T) {
		res := s.request(http.MethodGet, "/api/v1/wallets/not-a-uuid", ownerID, "")

		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestWalletHTTPRestrictions(t *testing.T) {
	s := newWalletHTTPTest(t)

	ownerID := s.createUser("customer", "active")
	adminID := s.createUser("admin", "active")

	w := s.createWallet(ownerID)

	walletPath := "/api/v1/wallets/" + w.ID.String()
	adminPath := "/api/v1/admin/wallets/" + w.ID.String()

	t.Run("owner freezes wallet", func(t *testing.T) {
		res := s.request(http.MethodPost, walletPath+"/freeze", ownerID, "")

		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("admin overrides owner freeze", func(t *testing.T) {
		res := s.request(
			http.MethodPost,
			adminPath+"/freeze",
			adminID,
			`{"reason":"Suspicious activity"}`,
		)

		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("owner cannot unfreeze admin restriction", func(t *testing.T) {
		res := s.request(
			http.MethodPost,
			walletPath+"/unfreeze",
			ownerID,
			"",
		)

		require.Equal(t, http.StatusConflict, res.Code)
	})

	t.Run("admin unfreezes wallet", func(t *testing.T) {
		res := s.request(
			http.MethodPost,
			adminPath+"/unfreeze",
			adminID,
			"",
		)

		require.Equal(t, http.StatusOK, res.Code)
	})

	found, err := s.repo.FindByID(
		context.Background(),
		w.ID,
		ownerID,
	)

	require.NoError(t, err)
	require.Equal(t, model.WalletStatusActive, found.Status)
	require.Nil(t, found.RestrictionType)
	require.Nil(t, found.RestrictedBy)
	require.Nil(t, found.RestrictedAt)
}

func TestWalletHTTPAdminAuthorization(t *testing.T) {
	s := newWalletHTTPTest(t)

	ownerID := s.createUser("customer", "active")
	adminID := s.createUser("admin", "active")

	w := s.createWallet(ownerID)

	path := "/api/v1/admin/wallets/" +
		w.ID.String() + "/freeze"

	body := `{"reason":"Security investigation"}`

	t.Run("customer denied admin access", func(t *testing.T) {
		res := s.request(http.MethodPost, path, ownerID, body)

		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("admin allowed", func(t *testing.T) {
		res := s.request(http.MethodPost, path, adminID, body)

		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("revoked admin denied", func(t *testing.T) {
		err := s.db.Exec(
			"UPDATE users SET role = ? WHERE id = ?",
			"customer",
			adminID,
		).Error

		require.NoError(t, err)

		res := s.request(http.MethodPost, path, adminID, body)

		require.Equal(t, http.StatusForbidden, res.Code)
	})
}
