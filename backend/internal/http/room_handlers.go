package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
)

type roomStorer interface {
	ListRooms(ctx context.Context) ([]model.Room, error)
	CreateRoom(ctx context.Context, id, name string, desc *string, cap *int) (model.Room, error)
}

type createRoomReq struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type roomResp struct {
	Room model.Room `json:"room"`
}

type roomsResp struct {
	Rooms []model.Room `json:"rooms"`
}

func CreateRoom(s roomStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin role required")
			return
		}

		var req createRoomReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required")
			return
		}

		room, err := s.CreateRoom(r.Context(), uuid.New().String(), req.Name, req.Description, req.Capacity)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create room")
			return
		}

		WriteJSON(w, http.StatusCreated, roomResp{Room: room})
	}
}

func ListRooms(s roomStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rooms, err := s.ListRooms(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list rooms")
			return
		}
		if rooms == nil {
			rooms = []model.Room{}
		}
		WriteJSON(w, http.StatusOK, roomsResp{Rooms: rooms})
	}
}
