package httpx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/avito-test/room-booking/internal/model"
)

type slotStorer interface {
	RoomExists(ctx context.Context, roomID string) (bool, error)
	GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error)
	UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error)
	SlotHasActiveBooking(ctx context.Context, slotID string) (bool, error)
}

type slotsResp struct {
	Slots []model.Slot `json:"slots"`
}

func ListSlots(s slotStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("roomId")
		dateStr := r.URL.Query().Get("date")
		if dateStr == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date parameter is required")
			return
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, use YYYY-MM-DD")
			return
		}

		exists, err := s.RoomExists(r.Context(), roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
			return
		}

		sch, err := s.GetSchedule(r.Context(), roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if sch == nil {
			WriteJSON(w, http.StatusOK, slotsResp{Slots: []model.Slot{}})
			return
		}

		dayOfWeek := int(date.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		dayAllowed := false
		for _, d := range sch.DaysOfWeek {
			if d == dayOfWeek {
				dayAllowed = true
				break
			}
		}
		if !dayAllowed {
			WriteJSON(w, http.StatusOK, slotsResp{Slots: []model.Slot{}})
			return
		}

		slots, err := generateSlots(r.Context(), s, roomID, date, sch)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate slots")
			return
		}

		available := slots[:0]
		for _, sl := range slots {
			booked, err := s.SlotHasActiveBooking(r.Context(), sl.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
				return
			}
			if !booked {
				available = append(available, sl)
			}
		}

		if available == nil {
			available = []model.Slot{}
		}
		WriteJSON(w, http.StatusOK, slotsResp{Slots: available})
	}
}

func generateSlots(ctx context.Context, s slotStorer, roomID string, date time.Time, sch *model.Schedule) ([]model.Slot, error) {
	sh, sm := parseTime(sch.StartTime)
	eh, em := parseTime(sch.EndTime)

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

	var slots []model.Slot
	for t := dayStart; t.Add(30*time.Minute).Before(dayEnd) || t.Add(30*time.Minute).Equal(dayEnd); t = t.Add(30 * time.Minute) {
		slotEnd := t.Add(30 * time.Minute)
		sl, err := s.UpsertSlot(ctx, roomID, t, slotEnd)
		if err != nil {
			return nil, fmt.Errorf("upsert slot: %w", err)
		}
		slots = append(slots, sl)
	}
	return slots, nil
}

type allSlotStorer interface {
	RoomExists(ctx context.Context, roomID string) (bool, error)
	GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error)
	UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error)
	GetAllSlotsWithBookings(ctx context.Context, roomID string, dateStart, dateEnd time.Time) ([]model.SlotWithBooking, error)
}

type allSlotsResp struct {
	Slots []model.SlotWithBooking `json:"slots"`
}

func ListAllSlots(s allSlotStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("roomId")
		dateStr := r.URL.Query().Get("date")
		if dateStr == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date parameter is required")
			return
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, use YYYY-MM-DD")
			return
		}

		exists, err := s.RoomExists(r.Context(), roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
			return
		}

		sch, err := s.GetSchedule(r.Context(), roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if sch == nil {
			WriteJSON(w, http.StatusOK, allSlotsResp{Slots: []model.SlotWithBooking{}})
			return
		}

		dayOfWeek := int(date.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		dayAllowed := false
		for _, d := range sch.DaysOfWeek {
			if d == dayOfWeek {
				dayAllowed = true
				break
			}
		}
		if !dayAllowed {
			WriteJSON(w, http.StatusOK, allSlotsResp{Slots: []model.SlotWithBooking{}})
			return
		}

		sh, sm := parseTime(sch.StartTime)
		eh, em := parseTime(sch.EndTime)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
		dayEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

		for t := dayStart; t.Add(30*time.Minute).Before(dayEnd) || t.Add(30*time.Minute).Equal(dayEnd); t = t.Add(30 * time.Minute) {
			slotEnd := t.Add(30 * time.Minute)
			if _, err := s.UpsertSlot(r.Context(), roomID, t, slotEnd); err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate slots")
				return
			}
		}

		slots, err := s.GetAllSlotsWithBookings(r.Context(), roomID, dayStart, dayEnd)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}

		WriteJSON(w, http.StatusOK, allSlotsResp{Slots: slots})
	}
}

func parseTime(s string) (int, int) {
	parts := strings.Split(s, ":")
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m
}
