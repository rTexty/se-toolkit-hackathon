package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/store"
)

type dummyLoginReq struct {
	Role string `json:"role"`
}

type tokenResp struct {
	Token string `json:"token"`
}

func DummyLogin(secret string, s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dummyLoginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		var userID, email string
		switch req.Role {
		case "admin":
			userID = auth.AdminID
			email = "admin@test.local"
		case "user":
			userID = auth.UserID
			email = "user@test.local"
		default:
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be admin or user")
			return
		}

		if s != nil {
			_ = s.EnsureUserExists(r.Context(), userID, email, req.Role)
		}

		token, err := auth.GenerateToken(secret, userID, req.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
			return
		}

		WriteJSON(w, http.StatusOK, tokenResp{Token: token})
	}
}
