package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type bookingStorer interface {
	GetSlot(ctx context.Context, slotID string) (*model.Slot, error)
	CreateBooking(ctx context.Context, id, slotID, userID string, confLink *string) (model.Booking, error)
	GetBookingByID(ctx context.Context, id string) (*model.Booking, error)
	CancelBooking(ctx context.Context, id string) (model.Booking, error)
	ListAllBookings(ctx context.Context, page, pageSize int) ([]model.Booking, int, error)
	ListUserBookings(ctx context.Context, userID string) ([]model.Booking, error)
}

type createBookingReq struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type bookingResp struct {
	Booking interface{} `json:"booking"`
}

type bookingsListResp struct {
	Bookings   interface{} `json:"bookings"`
	Pagination *pageResp   `json:"pagination,omitempty"`
}

type pageResp struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func CreateBooking(secret string, s bookingStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "user" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "booking is only available for users")
			return
		}

		var req createBookingReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid body")
			return
		}
		if req.SlotID == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "slotId is required")
			return
		}

		slot, err := s.GetSlot(r.Context(), req.SlotID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if slot == nil {
			writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
			return
		}

		slotStart, err := time.Parse(time.RFC3339, slot.Start)
		if err != nil {
			slotStart, err = time.Parse("2006-01-02 15:04:05-07", slot.Start)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "bad slot time")
			return
		}
		if slotStart.Before(time.Now().UTC()) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "cannot book a slot in the past")
			return
		}

		var confLink *string
		if req.CreateConferenceLink {
			link := "https://meet.example.com/" + uuid.New().String()[:8]
			confLink = &link
		}

		b, err := s.CreateBooking(r.Context(), uuid.New().String(), req.SlotID, claims.UserID, confLink)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				writeError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
				return
			}
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create booking")
			return
		}

		WriteJSON(w, http.StatusCreated, bookingResp{Booking: b})
	}
}

func CancelBooking(s bookingStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "user" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "only users can cancel bookings")
			return
		}

		bookingID := r.PathValue("bookingId")
		b, err := s.GetBookingByID(r.Context(), bookingID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if b == nil {
			writeError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
			return
		}
		if b.UserID != claims.UserID {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
			return
		}

		if b.Status == "cancelled" {
			WriteJSON(w, http.StatusOK, bookingResp{Booking: *b})
			return
		}

		cancelled, err := s.CancelBooking(r.Context(), bookingID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel")
			return
		}

		WriteJSON(w, http.StatusOK, bookingResp{Booking: cancelled})
	}
}

func ListAllBookings(s bookingStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin role required")
			return
		}

		page := 1
		pageSize := 20
		if p := r.URL.Query().Get("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v >= 1 {
				page = v
			}
		}
		if ps := r.URL.Query().Get("pageSize"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v >= 1 && v <= 100 {
				pageSize = v
			}
		}

		bookings, total, err := s.ListAllBookings(r.Context(), page, pageSize)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if bookings == nil {
			bookings = []model.Booking{}
		}

		WriteJSON(w, http.StatusOK, bookingsListResp{
			Bookings:   bookings,
			Pagination: &pageResp{Page: page, PageSize: pageSize, Total: total},
		})
	}
}

func ListMyBookings(s bookingStorer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims.Role != "user" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "only users can view their bookings")
			return
		}

		bookings, err := s.ListUserBookings(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db error")
			return
		}
		if bookings == nil {
			bookings = []model.Booking{}
		}

		WriteJSON(w, http.StatusOK, bookingsListResp{Bookings: bookings})
	}
}
