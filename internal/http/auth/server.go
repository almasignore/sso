package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"sso/internal/storage"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	auth Auth
}

func New(auth Auth) http.Handler {
	s := &serverAPI{
		auth: auth,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", s.Register)
	mux.HandleFunc("POST /login", s.Login)
	mux.HandleFunc("GET /is-admin", s.IsAdmin)

	return mux
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID int64 `json:"user_id"`
}

func (s *serverAPI) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	userID, err := s.auth.RegisterNewUser(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, RegisterResponse{
		UserID: userID,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AppID    int    `json:"app_id"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (s *serverAPI) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.AppID <= 0 {
		http.Error(w, "email, password and app_id are required", http.StatusBadRequest)
		return
	}

	token, err := s.auth.Login(
		r.Context(),
		req.Email,
		req.Password,
		req.AppID,
	)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{
		Token: token,
	})
}

type IsAdminResponse struct {
	IsAdmin bool `json:"is_admin"`
}

func (s *serverAPI) IsAdmin(
	w http.ResponseWriter,
	r *http.Request,
) {
	userIDStr := r.URL.Query().Get("user_id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	isAdmin, err := s.auth.IsAdmin(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, IsAdminResponse{
		IsAdmin: isAdmin,
	})
}

func writeJSON(
	w http.ResponseWriter,
	statusCode int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}
