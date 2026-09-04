package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

// AuthStatusResponse defines payload for GET /api/auth/status.
type AuthStatusResponse struct {
	SetupRequired bool              `json:"setup_required"`
	Authenticated bool              `json:"authenticated"`
	User          *auth.CurrentUser `json:"user,omitempty"`
}

// SetupRequest defines payload for POST /api/auth/setup.
type SetupRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginRequest defines payload for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateUserRequest defines payload for POST /api/users.
type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
}

// UserItem defines representation of a user item.
type UserItem struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

// AuthMiddleware inspects session cookie or Authorization header.
func (c *RouterConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var token string
		if cookie, err := r.Cookie(auth.CookieName); err == nil && cookie.Value != "" {
			token = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token != "" {
			session, err := c.DB.GetSessionWithUser(ctx, token)
			if err == nil {
				currentUser := &auth.CurrentUser{
					ID:          session.UserID,
					Username:    session.Username,
					Role:        session.Role,
					DisplayName: session.DisplayName,
				}
				ctx = auth.WithUser(ctx, currentUser)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth ensures the request has an authenticated user.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized: please log in")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin ensures the authenticated user is an administrator.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil || user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "forbidden: administrator privileges required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterAuthRoutes mounts auth-related routes.
func (c *RouterConfig) RegisterAuthRoutes(r chi.Router) {
	r.Route("/auth", func(authGroup chi.Router) {
		// GET /api/auth/status
		authGroup.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			count, err := c.DB.CountUsers(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to check system status")
				return
			}

			setupRequired := count == 0
			currentUser := auth.GetUser(r.Context())

			writeJSON(w, http.StatusOK, AuthStatusResponse{
				SetupRequired: setupRequired,
				Authenticated: currentUser != nil,
				User:          currentUser,
			})
		})

		// POST /api/auth/setup
		authGroup.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
			count, err := c.DB.CountUsers(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to check user count")
				return
			}
			if count > 0 {
				writeError(w, http.StatusForbidden, "first-run setup has already been completed")
				return
			}

			var req SetupRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			req.Username = strings.TrimSpace(req.Username)
			if req.Username == "" || len(req.Password) < 6 {
				writeError(w, http.StatusBadRequest, "username required and password must be at least 6 characters")
				return
			}

			hash, err := auth.HashPassword(req.Password)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to hash password")
				return
			}

			displayName := strings.TrimSpace(req.DisplayName)
			if displayName == "" {
				displayName = req.Username
			}

			user, err := c.DB.CreateUser(r.Context(), database.CreateUserParams{
				Username:     req.Username,
				PasswordHash: hash,
				Role:         auth.RoleAdmin,
				DisplayName:  displayName,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create admin user")
				return
			}

			// Create initial session
			token, err := auth.GenerateToken()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate session token")
				return
			}

			expiresAt := time.Now().Add(auth.SessionDuration)
			if _, err := c.DB.CreateSession(r.Context(), database.CreateSessionParams{
				Token:     token,
				UserID:    user.ID,
				ExpiresAt: expiresAt,
			}); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create session")
				return
			}

			auth.SetSessionCookie(w, token, expiresAt)

			writeJSON(w, http.StatusCreated, map[string]any{
				"message": "admin account created successfully",
				"user": auth.CurrentUser{
					ID:          user.ID,
					Username:    user.Username,
					Role:        user.Role,
					DisplayName: user.DisplayName,
				},
			})
		})

		// POST /api/auth/login
		authGroup.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			var req LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			req.Username = strings.TrimSpace(req.Username)
			if req.Username == "" || req.Password == "" {
				writeError(w, http.StatusBadRequest, "username and password required")
				return
			}

			user, err := c.DB.GetUserByUsername(r.Context(), req.Username)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusUnauthorized, "invalid username or password")
					return
				}
				writeError(w, http.StatusInternalServerError, "database error")
				return
			}

			if !auth.CheckPassword(user.PasswordHash, req.Password) {
				writeError(w, http.StatusUnauthorized, "invalid username or password")
				return
			}

			token, err := auth.GenerateToken()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate session token")
				return
			}

			expiresAt := time.Now().Add(auth.SessionDuration)
			if _, err := c.DB.CreateSession(r.Context(), database.CreateSessionParams{
				Token:     token,
				UserID:    user.ID,
				ExpiresAt: expiresAt,
			}); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create session")
				return
			}

			auth.SetSessionCookie(w, token, expiresAt)

			writeJSON(w, http.StatusOK, map[string]any{
				"message": "login successful",
				"user": auth.CurrentUser{
					ID:          user.ID,
					Username:    user.Username,
					Role:        user.Role,
					DisplayName: user.DisplayName,
				},
			})
		})

		// POST /api/auth/logout
		authGroup.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			if cookie, err := r.Cookie(auth.CookieName); err == nil && cookie.Value != "" {
				_ = c.DB.DeleteSession(r.Context(), cookie.Value)
			}
			auth.ClearSessionCookie(w)
			writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		})

		// GET /api/auth/me
		authGroup.With(RequireAuth).Get("/me", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUser(r.Context())
			writeJSON(w, http.StatusOK, user)
		})
	})

	// User management endpoints (Admin only)
	r.Route("/users", func(usersGroup chi.Router) {
		usersGroup.Use(RequireAuth)
		usersGroup.Use(RequireAdmin)

		// GET /api/users
		usersGroup.Get("/", func(w http.ResponseWriter, r *http.Request) {
			dbUsers, err := c.DB.ListUsers(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to list users")
				return
			}

			items := make([]UserItem, 0, len(dbUsers))
			for _, u := range dbUsers {
				items = append(items, UserItem{
					ID:          u.ID,
					Username:    u.Username,
					Role:        u.Role,
					DisplayName: u.DisplayName,
					CreatedAt:   formatTime(u.CreatedAt),
				})
			}

			writeJSON(w, http.StatusOK, items)
		})

		// POST /api/users
		usersGroup.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var req CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			req.Username = strings.TrimSpace(req.Username)
			if req.Username == "" || len(req.Password) < 6 {
				writeError(w, http.StatusBadRequest, "username required and password must be at least 6 characters")
				return
			}

			role := strings.ToLower(strings.TrimSpace(req.Role))
			if role != auth.RoleAdmin && role != auth.RoleReader {
				role = auth.RoleReader
			}

			hash, err := auth.HashPassword(req.Password)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to hash password")
				return
			}

			displayName := strings.TrimSpace(req.DisplayName)
			if displayName == "" {
				displayName = req.Username
			}

			user, err := c.DB.CreateUser(r.Context(), database.CreateUserParams{
				Username:     req.Username,
				PasswordHash: hash,
				Role:         role,
				DisplayName:  displayName,
			})
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					writeError(w, http.StatusConflict, "username already exists")
					return
				}
				writeError(w, http.StatusInternalServerError, "failed to create user")
				return
			}

			writeJSON(w, http.StatusCreated, UserItem{
				ID:          user.ID,
				Username:    user.Username,
				Role:        user.Role,
				DisplayName: user.DisplayName,
				CreatedAt:   formatTime(user.CreatedAt),
			})
		})

		// DELETE /api/users/{id}
		usersGroup.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid user id")
				return
			}

			currentUser := auth.GetUser(r.Context())
			if currentUser != nil && currentUser.ID == targetID {
				writeError(w, http.StatusBadRequest, "cannot delete currently logged in account")
				return
			}

			if err := c.DB.DeleteUser(r.Context(), targetID); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to delete user")
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
		})
	})
}
