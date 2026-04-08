package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var timeRe = regexp.MustCompile(`^([01]?\d|2[0-3]):[0-5]\d$`)

type scheduleStorer interface {
	RoomExists(ctx context.Context, roomID string) (bool, error)
	ScheduleExists(ctx context.Context, roomID string) (bool, error)
	CreateSchedule(ctx context.Context, id, roomID string, days []int, start, end string) (model.Schedule, error)
}

type scheduleReq struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type scheduleResp struct {
	Schedule interface{} `json:"schedule"`
}

func CreateSchedule(s scheduleStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin role required")
			return
		}

		roomID := r.PathValue("roomId")
		exists, err := s.RoomExists(r.Context(), roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
			return
		}

		already, _ := s.ScheduleExists(r.Context(), roomID)
		if already {
			writeError(w, http.StatusConflict, "SCHEDULE_EXISTS",
				"schedule for this room already exists and cannot be changed")
			return
		}

		var req scheduleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid body")
			return
		}

		if len(req.DaysOfWeek) == 0 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek must not be empty")
			return
		}
		for _, d := range req.DaysOfWeek {
			if d < 1 || d > 7 {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek values must be 1-7")
				return
			}
		}
		if !timeRe.MatchString(req.StartTime) || !timeRe.MatchString(req.EndTime) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid time format, use HH:MM")
			return
		}
		if req.StartTime >= req.EndTime {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "startTime must be before endTime")
			return
		}

		sch, err := s.CreateSchedule(r.Context(), uuid.New().String(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				writeError(w, http.StatusConflict, "SCHEDULE_EXISTS", "schedule already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create schedule")
			return
		}

		WriteJSON(w, http.StatusCreated, scheduleResp{Schedule: sch})
	}
}
