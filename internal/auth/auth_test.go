package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPasswordHashingAndChecking(t *testing.T) {
	password := "SecretP@ss123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Errorf("expected CheckPassword to return true for matching password")
	}

	if CheckPassword(hash, "WrongPassword") {
		t.Errorf("expected CheckPassword to return false for wrong password")
	}
}

func TestGenerateToken(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if len(token1) != 64 {
		t.Errorf("expected token length 64, got %d", len(token1))
	}

	token2, _ := GenerateToken()
	if token1 == token2 {
		t.Errorf("expected tokens to be unique, got duplicate")
	}
}

func TestCookieHelpers(t *testing.T) {
	rec := httptest.NewRecorder()
	SetSessionCookie(rec, "dummy-token-abc", time.Now().Add(1*time.Hour))

	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie to be set")
	}

	cookie := cookies[0]
	if cookie.Name != CookieName || cookie.Value != "dummy-token-abc" {
		t.Errorf("unexpected cookie: %+v", cookie)
	}
	if !cookie.HttpOnly {
		t.Errorf("expected HttpOnly cookie")
	}

	rec2 := httptest.NewRecorder()
	ClearSessionCookie(rec2)
	res2 := rec2.Result()
	cookies2 := res2.Cookies()
	if len(cookies2) == 0 || cookies2[0].MaxAge >= 0 {
		t.Errorf("expected cleared cookie with negative MaxAge")
	}
}

func TestUserContext(t *testing.T) {
	user := &CurrentUser{
		ID:          42,
		Username:    "admin",
		Role:        RoleAdmin,
		DisplayName: "Site Administrator",
	}

	ctx := WithUser(context.Background(), user)
	extracted := GetUser(ctx)
	if extracted == nil || extracted.ID != 42 || extracted.Username != "admin" {
		t.Errorf("expected extracted user to match, got: %+v", extracted)
	}

	nilUser := GetUser(context.Background())
	if nilUser != nil {
		t.Errorf("expected nil user from empty context")
	}
}
