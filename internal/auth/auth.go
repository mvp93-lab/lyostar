package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	CookieName      = "lyostar_session"
	SessionDuration = 30 * 24 * time.Hour

	RoleAdmin  = "admin"
	RoleReader = "reader"
)

type contextKey string

const userContextKey contextKey = "lyostar_user"

// Permissions represents granular capabilities assigned to a user.
type Permissions struct {
	CanRead     bool `json:"can_read"`
	CanDownload bool `json:"can_download"`
	CanUpload   bool `json:"can_upload"`
	CanEdit     bool `json:"can_edit"`
	CanDelete   bool `json:"can_delete"`
}

// CurrentUser represents authenticated user info in context.
type CurrentUser struct {
	ID          int64       `json:"id"`
	Username    string      `json:"username"`
	Role        string      `json:"role"`
	DisplayName string      `json:"display_name"`
	Permissions Permissions `json:"permissions"`
}

// HashPassword hashes a plain text password with bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a bcrypt hashed password with its plain representation.
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken generates a cryptographically secure 64-character hex string.
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SetSessionCookie configures a secure, HTTP-only session cookie.
func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie clears the session cookie from the client.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// WithUser adds the CurrentUser to the context.
func WithUser(ctx context.Context, u *CurrentUser) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// GetUser retrieves the CurrentUser from the context if present.
func GetUser(ctx context.Context) *CurrentUser {
	u, ok := ctx.Value(userContextKey).(*CurrentUser)
	if !ok {
		return nil
	}
	return u
}
