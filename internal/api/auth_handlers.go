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
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	Role        string            `json:"role"`
	DisplayName string            `json:"display_name"`
	Permissions *auth.Permissions `json:"permissions,omitempty"`
}

// UpdateUserRequest defines payload for PUT /api/users/{id}.
type UpdateUserRequest struct {
	DisplayName string            `json:"display_name"`
	Role        string            `json:"role"`
	Password    string            `json:"password,omitempty"`
	Permissions *auth.Permissions `json:"permissions,omitempty"`
}

// UserItem defines representation of a user item.
type UserItem struct {
	ID          int64            `json:"id"`
	Username    string           `json:"username"`
	Role        string           `json:"role"`
	DisplayName string           `json:"display_name"`
	Permissions auth.Permissions `json:"permissions"`
	CreatedAt   string           `json:"created_at"`
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
					Permissions: auth.Permissions{
						CanRead:     session.CanRead == 1,
						CanDownload: session.CanDownload == 1,
						CanUpload:   session.CanUpload == 1,
						CanEdit:     session.CanEdit == 1,
						CanDelete:   session.CanDelete == 1,
					},
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

// RequirePermission checks whether the authenticated user has a specific permission.
func RequirePermission(perm func(p auth.Permissions) bool, actionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUser(r.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized: please log in")
				return
			}
			if !perm(user.Permissions) {
				writeError(w, http.StatusForbidden, "forbidden: "+actionName+" permission required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRead ensures user has can_read capability.
func RequireRead(next http.Handler) http.Handler {
	return RequirePermission(func(p auth.Permissions) bool { return p.CanRead }, "read")(next)
}

// RequireDownload ensures user has can_download capability.
func RequireDownload(next http.Handler) http.Handler {
	return RequirePermission(func(p auth.Permissions) bool { return p.CanDownload }, "download")(next)
}

// RequireUpload ensures user has can_upload capability.
func RequireUpload(next http.Handler) http.Handler {
	return RequirePermission(func(p auth.Permissions) bool { return p.CanUpload }, "upload")(next)
}

// RequireEdit ensures user has can_edit capability.
func RequireEdit(next http.Handler) http.Handler {
	return RequirePermission(func(p auth.Permissions) bool { return p.CanEdit }, "edit")(next)
}

// RequireDelete ensures user has can_delete capability.
func RequireDelete(next http.Handler) http.Handler {
	return RequirePermission(func(p auth.Permissions) bool { return p.CanDelete }, "delete")(next)
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
				CanRead:      1,
				CanDownload:  1,
				CanUpload:    1,
				CanEdit:      1,
				CanDelete:    1,
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
					Permissions: auth.Permissions{
						CanRead:     user.CanRead == 1,
						CanDownload: user.CanDownload == 1,
						CanUpload:   user.CanUpload == 1,
						CanEdit:     user.CanEdit == 1,
						CanDelete:   user.CanDelete == 1,
					},
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
					Permissions: auth.Permissions{
						CanRead:     user.CanRead == 1,
						CanDownload: user.CanDownload == 1,
						CanUpload:   user.CanUpload == 1,
						CanEdit:     user.CanEdit == 1,
						CanDelete:   user.CanDelete == 1,
					},
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
					Permissions: auth.Permissions{
						CanRead:     u.CanRead == 1,
						CanDownload: u.CanDownload == 1,
						CanUpload:   u.CanUpload == 1,
						CanEdit:     u.CanEdit == 1,
						CanDelete:   u.CanDelete == 1,
					},
					CreatedAt: formatTime(u.CreatedAt),
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

			var canRead, canDownload, canUpload, canEdit, canDelete int64
			if req.Permissions != nil {
				if req.Permissions.CanRead { canRead = 1 }
				if req.Permissions.CanDownload { canDownload = 1 }
				if req.Permissions.CanUpload { canUpload = 1 }
				if req.Permissions.CanEdit { canEdit = 1 }
				if req.Permissions.CanDelete { canDelete = 1 }
			} else {
				// Default permissions: Readers can read & download. Admins have all permissions.
				canRead = 1
				canDownload = 1
				if role == auth.RoleAdmin {
					canUpload = 1
					canEdit = 1
					canDelete = 1
				}
			}

			user, err := c.DB.CreateUser(r.Context(), database.CreateUserParams{
				Username:     req.Username,
				PasswordHash: hash,
				Role:         role,
				DisplayName:  displayName,
				CanRead:      canRead,
				CanDownload:  canDownload,
				CanUpload:    canUpload,
				CanEdit:      canEdit,
				CanDelete:    canDelete,
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
				Permissions: auth.Permissions{
					CanRead:     user.CanRead == 1,
					CanDownload: user.CanDownload == 1,
					CanUpload:   user.CanUpload == 1,
					CanEdit:     user.CanEdit == 1,
					CanDelete:   user.CanDelete == 1,
				},
				CreatedAt: formatTime(user.CreatedAt),
			})
		})

		// PUT /api/users/{id}
		usersGroup.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
			targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid user id")
				return
			}

			existing, err := c.DB.GetUserByID(r.Context(), targetID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusNotFound, "user not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "database error")
				return
			}

			var req UpdateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			displayName := strings.TrimSpace(req.DisplayName)
			if displayName == "" {
				displayName = existing.DisplayName
			}

			role := strings.ToLower(strings.TrimSpace(req.Role))
			if role == "" {
				role = existing.Role
			} else if role != auth.RoleAdmin && role != auth.RoleReader {
				role = auth.RoleReader
			}

			// Prevent demoting the last admin if this user is the only admin
			if existing.Role == auth.RoleAdmin && role != auth.RoleAdmin {
				usersList, _ := c.DB.ListUsers(r.Context())
				adminCount := 0
				for _, u := range usersList {
					if u.Role == auth.RoleAdmin {
						adminCount++
					}
				}
				if adminCount <= 1 {
					writeError(w, http.StatusBadRequest, "cannot demote the sole administrator")
					return
				}
			}

			canRead := existing.CanRead
			canDownload := existing.CanDownload
			canUpload := existing.CanUpload
			canEdit := existing.CanEdit
			canDelete := existing.CanDelete

			if req.Permissions != nil {
				if req.Permissions.CanRead { canRead = 1 } else { canRead = 0 }
				if req.Permissions.CanDownload { canDownload = 1 } else { canDownload = 0 }
				if req.Permissions.CanUpload { canUpload = 1 } else { canUpload = 0 }
				if req.Permissions.CanEdit { canEdit = 1 } else { canEdit = 0 }
				if req.Permissions.CanDelete { canDelete = 1 } else { canDelete = 0 }
			}

			updated, err := c.DB.UpdateUser(r.Context(), database.UpdateUserParams{
				ID:          targetID,
				DisplayName: displayName,
				Role:        role,
				CanRead:     canRead,
				CanDownload: canDownload,
				CanUpload:   canUpload,
				CanEdit:     canEdit,
				CanDelete:   canDelete,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update user")
				return
			}

			// If password was provided and not empty, update password
			if strings.TrimSpace(req.Password) != "" {
				if len(req.Password) < 6 {
					writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
					return
				}
				newHash, err := auth.HashPassword(req.Password)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "failed to hash password")
					return
				}
				_, err = c.DB.UpdateUserPassword(r.Context(), database.UpdateUserPasswordParams{
					ID:           targetID,
					PasswordHash: newHash,
				})
				if err != nil {
					writeError(w, http.StatusInternalServerError, "failed to update password")
					return
				}
			}

			writeJSON(w, http.StatusOK, UserItem{
				ID:          updated.ID,
				Username:    updated.Username,
				Role:        updated.Role,
				DisplayName: updated.DisplayName,
				Permissions: auth.Permissions{
					CanRead:     updated.CanRead == 1,
					CanDownload: updated.CanDownload == 1,
					CanUpload:   updated.CanUpload == 1,
					CanEdit:     updated.CanEdit == 1,
					CanDelete:   updated.CanDelete == 1,
				},
				CreatedAt: formatTime(updated.CreatedAt),
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
