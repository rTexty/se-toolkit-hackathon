package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/model"
	"github.com/avito-test/room-booking/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResp struct {
	User model.User `json:"user"`
}

func Register(secret string, s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
			return
		}
		if req.Role != "admin" && req.Role != "user" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be admin or user")
			return
		}
		if len(req.Password) < 6 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "password must be at least 6 characters")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password")
			return
		}

		user, err := s.CreateUser(r.Context(), uuid.New().String(), req.Email, string(hash), req.Role)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
			return
		}

		token, err := auth.GenerateToken(secret, user.ID, user.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
			return
		}

		w.Header().Set("X-Auth-Token", token)
		WriteJSON(w, http.StatusCreated, userResp{User: user})
	}
}

func Login(secret string, s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
			return
		}

		user, err := s.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if user == nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			return
		}

		token, err := auth.GenerateToken(secret, user.ID, user.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
			return
		}

		WriteJSON(w, http.StatusOK, tokenResp{Token: token})
	}
}
