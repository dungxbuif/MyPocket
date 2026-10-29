package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type googleStoreFailureRepo struct{ err error }

func (r googleStoreFailureRepo) FindByEmail(string) (*entity.User, error) { return nil, r.err }
func (r googleStoreFailureRepo) FindByID(string) (*entity.User, error)    { return nil, r.err }
func (r googleStoreFailureRepo) FindOrCreateGoogleUser(entity.GoogleProfile) (*entity.User, error) {
	return nil, r.err
}
func (r googleStoreFailureRepo) SetTimezone(string, string, bool) (*entity.User, error) {
	return nil, r.err
}
func (r googleStoreFailureRepo) Create(*entity.User) error { return r.err }

var _ repository.UserRepository = googleStoreFailureRepo{}

func TestMobileGoogleExchangeReturnsServiceUnavailableForUserStoreFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := usecase.NewAuthInteractor(
		googleStoreFailureRepo{err: errors.New("database connection refused")},
		nil,
		nil,
		nil,
		nil,
	)
	handler := NewAuthHandler(service, true, nil)
	router := gin.New()
	router.POST("/api/v1/auth/mobile/google/exchange", handler.MobileGoogleExchange)

	requestBody := []byte(`{"code":"fixture","code_verifier":"fixture","redirect_uri":"mypocket://oauth/callback","platform":"ios"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mobile/google/exchange", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var problem Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem.Code != problemCodeGoogleUserStoreUnavailable {
		t.Fatalf("problem code = %q, want %q", problem.Code, problemCodeGoogleUserStoreUnavailable)
	}
	if problem.Detail == "database connection refused" || problem.Detail == "" {
		t.Fatalf("response leaked or omitted persistence detail: %q", problem.Detail)
	}
}
