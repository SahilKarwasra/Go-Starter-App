package handler_test

import (
	"bytes"
	"context"
	"database/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api/cmd/server/handler"
	"api/cmd/server/routes"
	"api/cmd/server/services"
	"api/cmd/server/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockUserRepo) UpdateRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	u, ok := m.users[userID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	u.RefreshToken = refreshToken
	return nil
}

func setupTestApp() (http.Handler, *mockUserRepo) {
	repo := newMockUserRepo()
	jwtSecret := "test-secret"
	jwtRefreshSecret := "test-refresh-secret"

	authService := services.NewAuthService(
		repo,
		jwtSecret,
		jwtRefreshSecret,
		15*time.Minute,
		24*time.Hour,
	)
	authHandler := handler.NewAuthHandler(authService)
	router := routes.SetupRouter(jwtSecret, authHandler)

	return router, repo
}

func TestAuthEndpoints(t *testing.T) {
	router, _ := setupTestApp()

	// 1. Test Sign Up
	signUpBody, _ := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "Password123!",
		"name":     "Sahil",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBuffer(signUpBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var signUpResp utils.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &signUpResp); err != nil {
		t.Fatalf("failed to decode APIResponse: %v", err)
	}
	if !signUpResp.IsSuccess {
		t.Errorf("expected isSuccess true")
	}

	// 2. Duplicate Sign Up
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBuffer(signUpBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate signup, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Sign In with invalid password
	invalidSignInBody, _ := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "WrongPassword",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(invalidSignInBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad password, got %d", w.Code)
	}

	// 4. Sign In with valid credentials
	signInBody, _ := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "Password123!",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(signInBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var signInResp struct {
		StatusCode int  `json:"statusCode"`
		IsSuccess  bool `json:"isSuccess"`
		Data       struct {
			AccessToken  string                `json:"accessToken"`
			RefreshToken string                `json:"refreshToken"`
			User         services.UserResponse `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &signInResp); err != nil {
		t.Fatalf("failed to decode signin response: %v", err)
	}

	accessToken := signInResp.Data.AccessToken
	refreshToken := signInResp.Data.RefreshToken

	// 5. Access Protected Route: /api/v1/auth/me without token -> 401
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth header, got %d", w.Code)
	}

	// 6. Access Protected Route: /api/v1/auth/me with valid Bearer token -> 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with auth header, got %d: %s", w.Code, w.Body.String())
	}

	// 7. Refresh token -> 200
	refreshBody, _ := json.Marshal(map[string]string{
		"refreshToken": refreshToken,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/refresh-token", bytes.NewBuffer(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on refresh, got %d: %s", w.Code, w.Body.String())
	}

	// 8. Logout
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on logout, got %d", w.Code)
	}

	// 9. After logout, refresh token should fail -> 401
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/refresh-token", bytes.NewBuffer(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout refresh, got %d", w.Code)
	}
}
