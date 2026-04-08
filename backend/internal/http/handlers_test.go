package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/model"
)

func setTestClaims(ctx context.Context, userID, role string) context.Context {
	return auth.WithClaims(ctx, &auth.Claims{
		UserID: userID,
		Role:   role,
	})
}

// --- Mock stores ---

type mockRoomStorer struct {
	listRooms  func() ([]model.Room, error)
	createRoom func(id, name string, desc *string, cap *int) (model.Room, error)
}

func (m *mockRoomStorer) ListRooms(ctx context.Context) ([]model.Room, error) {
	return m.listRooms()
}

func (m *mockRoomStorer) CreateRoom(ctx context.Context, id, name string, desc *string, cap *int) (model.Room, error) {
	return m.createRoom(id, name, desc, cap)
}

type mockScheduleStorer struct {
	roomExists     func() (bool, error)
	scheduleExists func() (bool, error)
	createSchedule func(id, roomID string, days []int, start, end string) (model.Schedule, error)
}

func (m *mockScheduleStorer) RoomExists(ctx context.Context, roomID string) (bool, error) {
	return m.roomExists()
}

func (m *mockScheduleStorer) ScheduleExists(ctx context.Context, roomID string) (bool, error) {
	return m.scheduleExists()
}

func (m *mockScheduleStorer) CreateSchedule(ctx context.Context, id, roomID string, days []int, start, end string) (model.Schedule, error) {
	return m.createSchedule(id, roomID, days, start, end)
}

type mockSlotStorer struct {
	roomExists           func() (bool, error)
	getSchedule          func() (*model.Schedule, error)
	upsertSlot           func(roomID string, start, end time.Time) (model.Slot, error)
	slotHasActiveBooking func(slotID string) (bool, error)
}

func (m *mockSlotStorer) RoomExists(ctx context.Context, roomID string) (bool, error) {
	return m.roomExists()
}

func (m *mockSlotStorer) GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error) {
	return m.getSchedule()
}

func (m *mockSlotStorer) UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error) {
	return m.upsertSlot(roomID, start, end)
}

func (m *mockSlotStorer) SlotHasActiveBooking(ctx context.Context, slotID string) (bool, error) {
	return m.slotHasActiveBooking(slotID)
}

type mockBookingStorer struct {
	getSlot          func() (*model.Slot, error)
	createBooking    func(id, slotID, userID string, confLink *string) (model.Booking, error)
	getBookingByID   func() (*model.Booking, error)
	cancelBooking    func(id string) (model.Booking, error)
	listAllBookings  func(page, pageSize int) ([]model.Booking, int, error)
	listUserBookings func(userID string) ([]model.Booking, error)
}

func (m *mockBookingStorer) GetSlot(ctx context.Context, slotID string) (*model.Slot, error) {
	return m.getSlot()
}

func (m *mockBookingStorer) CreateBooking(ctx context.Context, id, slotID, userID string, confLink *string) (model.Booking, error) {
	return m.createBooking(id, slotID, userID, confLink)
}

func (m *mockBookingStorer) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	return m.getBookingByID()
}

func (m *mockBookingStorer) CancelBooking(ctx context.Context, id string) (model.Booking, error) {
	return m.cancelBooking(id)
}

func (m *mockBookingStorer) ListAllBookings(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	return m.listAllBookings(page, pageSize)
}

func (m *mockBookingStorer) ListUserBookings(ctx context.Context, userID string) ([]model.Booking, error) {
	return m.listUserBookings(userID)
}

// --- TestListRooms ---

func TestListRoomsReturnsRooms(t *testing.T) {
	store := &mockRoomStorer{
		listRooms: func() ([]model.Room, error) {
			return []model.Room{{ID: "r1", Name: "Room 1"}}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListRooms(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp roomsResp
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Rooms) != 1 || resp.Rooms[0].Name != "Room 1" {
		t.Errorf("unexpected rooms: %+v", resp.Rooms)
	}
}

func TestListRoomsReturnsEmptySlice(t *testing.T) {
	store := &mockRoomStorer{
		listRooms: func() ([]model.Room, error) {
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListRooms(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	rooms, _ := resp["rooms"].([]interface{})
	if rooms == nil {
		t.Error("expected empty slice, got nil")
	}
}

func TestListRoomsInternalError(t *testing.T) {
	store := &mockRoomStorer{
		listRooms: func() ([]model.Room, error) {
			return nil, errors.New("db error")
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListRooms(store)(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- TestCreateRoom ---

func TestCreateRoomSuccess(t *testing.T) {
	store := &mockRoomStorer{
		createRoom: func(id, name string, desc *string, cap *int) (model.Room, error) {
			return model.Room{ID: id, Name: name}, nil
		},
	}

	w := httptest.NewRecorder()
	body := `{"name":"New Room","description":"desc","capacity":10}`
	r := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateRoom(store)(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp roomResp
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Room.Name != "New Room" {
		t.Errorf("room name = %q, want New Room", resp.Room.Name)
	}
}

func TestCreateRoomInternalError(t *testing.T) {
	store := &mockRoomStorer{
		createRoom: func(id, name string, desc *string, cap *int) (model.Room, error) {
			return model.Room{}, errors.New("db error")
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString(`{"name":"Room"}`))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateRoom(store)(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- TestCreateSchedule validation ---

func TestCreateScheduleEmptyDaysOfWeek(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[],"startTime":"09:00","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleInvalidDayValue(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[0],"startTime":"09:00","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleInvalidTimeFormat(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1],"startTime":"9am","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleStartGreaterOrEqualEnd(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1],"startTime":"18:00","endTime":"09:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleStartEqualsEnd(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1],"startTime":"09:00","endTime":"09:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleRoomNotFound(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1],"startTime":"09:00","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCreateScheduleAlreadyExists(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return true, nil },
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1],"startTime":"09:00","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestCreateScheduleSuccess(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
		createSchedule: func(id, roomID string, days []int, start, end string) (model.Schedule, error) {
			return model.Schedule{ID: id, RoomID: roomID, DaysOfWeek: days, StartTime: start, EndTime: end}, nil
		},
	}

	w := httptest.NewRecorder()
	body := `{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestCreateScheduleInvalidBody(t *testing.T) {
	store := &mockScheduleStorer{
		roomExists:     func() (bool, error) { return true, nil },
		scheduleExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/r1/schedule/create", bytes.NewBufferString("invalid"))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateSchedule(store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- TestListSlots ---

func TestListSlotsNoSchedule(t *testing.T) {
	store := &mockSlotStorer{
		roomExists:  func() (bool, error) { return true, nil },
		getSchedule: func() (*model.Schedule, error) { return nil, nil },
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/r1/slots/list?date=2026-04-06", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListSlots(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp slotsResp
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Slots) != 0 {
		t.Errorf("expected 0 slots, got %d", len(resp.Slots))
	}
}

func TestListSlotsDayNotAllowed(t *testing.T) {
	sch := &model.Schedule{
		ID:         "s1",
		RoomID:     "r1",
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	store := &mockSlotStorer{
		roomExists:  func() (bool, error) { return true, nil },
		getSchedule: func() (*model.Schedule, error) { return sch, nil },
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/r1/slots/list?date=2026-04-05", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListSlots(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp slotsResp
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Slots) != 0 {
		t.Errorf("expected 0 slots for non-working day, got %d", len(resp.Slots))
	}
}

func TestListSlotsRoomNotFound(t *testing.T) {
	store := &mockSlotStorer{
		roomExists: func() (bool, error) { return false, nil },
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/r1/slots/list?date=2026-04-06", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListSlots(store)(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- TestCancelBooking ---

func TestCancelBookingNotFound(t *testing.T) {
	store := &mockBookingStorer{
		getBookingByID: func() (*model.Booking, error) { return nil, nil },
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/b1/cancel", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CancelBooking(store)(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCancelBookingAlreadyCancelled(t *testing.T) {
	store := &mockBookingStorer{
		getBookingByID: func() (*model.Booking, error) {
			return &model.Booking{ID: "b1", UserID: "user-1", Status: "cancelled"}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/b1/cancel", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CancelBooking(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp bookingResp
	json.NewDecoder(w.Body).Decode(&resp)
	b, _ := resp.Booking.(map[string]interface{})
	if b["status"] != "cancelled" {
		t.Errorf("status = %v, want cancelled", b["status"])
	}
}

func TestCancelBookingNotOwner(t *testing.T) {
	store := &mockBookingStorer{
		getBookingByID: func() (*model.Booking, error) {
			return &model.Booking{ID: "b1", UserID: "user-2", Status: "active"}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/b1/cancel", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CancelBooking(store)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCancelBookingSuccess(t *testing.T) {
	store := &mockBookingStorer{
		getBookingByID: func() (*model.Booking, error) {
			return &model.Booking{ID: "b1", UserID: "user-1", Status: "active"}, nil
		},
		cancelBooking: func(id string) (model.Booking, error) {
			return model.Booking{ID: id, UserID: "user-1", Status: "cancelled"}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/b1/cancel", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CancelBooking(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- TestListAllBookings with pagination ---

func TestListAllBookingsDefaultPagination(t *testing.T) {
	store := &mockBookingStorer{
		listAllBookings: func(page, pageSize int) ([]model.Booking, int, error) {
			if page != 1 || pageSize != 20 {
				t.Errorf("unexpected pagination: page=%d, pageSize=%d", page, pageSize)
			}
			return []model.Booking{}, 0, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	ListAllBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestListAllBookingsCustomPagination(t *testing.T) {
	store := &mockBookingStorer{
		listAllBookings: func(page, pageSize int) ([]model.Booking, int, error) {
			if page != 2 || pageSize != 5 {
				t.Errorf("unexpected pagination: page=%d, pageSize=%d", page, pageSize)
			}
			return []model.Booking{{ID: "b1"}}, 1, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/list?page=2&pageSize=5", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	ListAllBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp bookingsListResp
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Pagination.Page != 2 {
		t.Errorf("page = %d, want 2", resp.Pagination.Page)
	}
	if resp.Pagination.PageSize != 5 {
		t.Errorf("pageSize = %d, want 5", resp.Pagination.PageSize)
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Pagination.Total)
	}
}

func TestListAllBookingsInvalidPagination(t *testing.T) {
	store := &mockBookingStorer{
		listAllBookings: func(page, pageSize int) ([]model.Booking, int, error) {
			if page != 1 || pageSize != 20 {
				t.Errorf("expected defaults, got page=%d, pageSize=%d", page, pageSize)
			}
			return []model.Booking{}, 0, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/list?page=-1&pageSize=200", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	ListAllBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestListAllBookingsReturnsEmptySlice(t *testing.T) {
	store := &mockBookingStorer{
		listAllBookings: func(page, pageSize int) ([]model.Booking, int, error) {
			return nil, 0, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	ListAllBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	bookings, _ := resp["bookings"].([]interface{})
	if bookings == nil {
		t.Error("expected empty slice, got nil")
	}
}

// --- TestListMyBookings ---

func TestListMyBookingsReturnsBookings(t *testing.T) {
	store := &mockBookingStorer{
		listUserBookings: func(userID string) ([]model.Booking, error) {
			if userID != "user-1" {
				t.Errorf("userID = %s, want user-1", userID)
			}
			return []model.Booking{{ID: "b1", UserID: "user-1", Status: "active"}}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/my", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListMyBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp bookingsListResp
	json.NewDecoder(w.Body).Decode(&resp)
	b, _ := resp.Bookings.([]interface{})
	if len(b) != 1 {
		t.Errorf("expected 1 booking, got %d", len(b))
	}
}

func TestListMyBookingsReturnsEmptySlice(t *testing.T) {
	store := &mockBookingStorer{
		listUserBookings: func(userID string) ([]model.Booking, error) {
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/my", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListMyBookings(store)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	bookings, _ := resp["bookings"].([]interface{})
	if bookings == nil {
		t.Error("expected empty slice, got nil")
	}
}

// --- TestCreateBooking with conference link ---

func TestCreateBookingWithConferenceLink(t *testing.T) {
	future := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	store := &mockBookingStorer{
		getSlot: func() (*model.Slot, error) {
			return &model.Slot{ID: "slot-1", Start: future}, nil
		},
		createBooking: func(id, slotID, userID string, confLink *string) (model.Booking, error) {
			if confLink == nil {
				t.Error("expected conference link, got nil")
			}
			return model.Booking{ID: id, SlotID: slotID, UserID: userID, ConferenceLink: confLink, Status: "active"}, nil
		},
	}

	w := httptest.NewRecorder()
	body := `{"slotId":"slot-1","createConferenceLink":true}`
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", store)(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp bookingResp
	json.NewDecoder(w.Body).Decode(&resp)
	b, _ := resp.Booking.(map[string]interface{})
	if b["conferenceLink"] == nil {
		t.Error("expected conference link in response")
	}
}

func TestCreateBookingWithoutConferenceLink(t *testing.T) {
	future := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	store := &mockBookingStorer{
		getSlot: func() (*model.Slot, error) {
			return &model.Slot{ID: "slot-1", Start: future}, nil
		},
		createBooking: func(id, slotID, userID string, confLink *string) (model.Booking, error) {
			if confLink != nil {
				t.Errorf("expected nil conference link, got %s", *confLink)
			}
			return model.Booking{ID: id, SlotID: slotID, UserID: userID, Status: "active"}, nil
		},
	}

	w := httptest.NewRecorder()
	body := `{"slotId":"slot-1","createConferenceLink":false}`
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", store)(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestCreateBookingSlotNotFound(t *testing.T) {
	store := &mockBookingStorer{
		getSlot: func() (*model.Slot, error) { return nil, nil },
	}

	w := httptest.NewRecorder()
	body := `{"slotId":"slot-1"}`
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", store)(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCreateBookingPastSlot(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	store := &mockBookingStorer{
		getSlot: func() (*model.Slot, error) {
			return &model.Slot{ID: "slot-1", Start: past}, nil
		},
	}

	w := httptest.NewRecorder()
	body := `{"slotId":"slot-1"}`
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(body))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", store)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "TEST_CODE", "test message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	err, _ := resp["error"].(map[string]interface{})
	if err["code"] != "TEST_CODE" {
		t.Errorf("code = %v, want TEST_CODE", err["code"])
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["key"] != "value" {
		t.Errorf("key = %v, want value", resp["key"])
	}
}

func TestDummyLoginInvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString("invalid"))
	DummyLogin("secret", nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDummyLoginInvalidRole(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"role":"invalid"}`
	r := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(body))
	DummyLogin("secret", nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDummyLoginAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"role":"admin"}`
	r := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(body))
	DummyLogin("secret", nil)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct{ Token string }
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}

	claims, err := auth.ParseToken("secret", resp.Token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != auth.AdminID {
		t.Errorf("UserID = %s, want %s", claims.UserID, auth.AdminID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %s, want admin", claims.Role)
	}
}

func TestDummyLoginUser(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"role":"user"}`
	r := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(body))
	DummyLogin("secret", nil)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct{ Token string }
	json.NewDecoder(w.Body).Decode(&resp)
	claims, _ := auth.ParseToken("secret", resp.Token)
	if claims.UserID != auth.UserID {
		t.Errorf("UserID = %s, want %s", claims.UserID, auth.UserID)
	}
}

func TestCreateRoomForbiddenForUser(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString(`{"name":"test"}`))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateRoom(nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCreateRoomInvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString("invalid"))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateRoom(nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateRoomEmptyName(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString(`{"name":""}`))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateRoom(nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateScheduleForbiddenForUser(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/rooms/test-id/schedule/create", bytes.NewBufferString(`{}`))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateSchedule(nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCreateBookingForbiddenForAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(`{}`))
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CreateBooking("secret", nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCancelBookingForbiddenForAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/test-id/cancel", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	CancelBooking(nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestListAllBookingsForbiddenForUser(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListAllBookings(nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestListMyBookingsForbiddenForAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/bookings/my", nil)
	r = r.WithContext(setTestClaims(r.Context(), "admin-1", "admin"))
	ListMyBookings(nil)(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestListSlotsMissingDate(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/test-id/slots/list", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListSlots(nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListSlotsInvalidDate(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/test-id/slots/list?date=invalid", nil)
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	ListSlots(nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateBookingInvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString("invalid"))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateBookingEmptySlotID(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/bookings/create", bytes.NewBufferString(`{"slotId":""}`))
	r = r.WithContext(setTestClaims(r.Context(), "user-1", "user"))
	CreateBooking("secret", nil)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
