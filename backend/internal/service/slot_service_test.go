package service

import (
	"context"
	"testing"
	"time"

	"github.com/avito-test/room-booking/internal/model"
)

// Mock RoomStore
type mockRoomStore struct {
	roomExists bool
	roomErr    error
}

func (m *mockRoomStore) RoomExists(ctx context.Context, roomID string) (bool, error) {
	return m.roomExists, m.roomErr
}

// Mock ScheduleStore
type mockScheduleStore struct {
	schedule *model.Schedule
	err      error
}

func (m *mockScheduleStore) GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error) {
	return m.schedule, m.err
}

// Mock SlotStore
type mockSlotStore struct {
	upsertedSlot model.Slot
	upsertErr    error
	hasBooking   bool
	bookingErr   error
}

func (m *mockSlotStore) UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error) {
	return m.upsertedSlot, m.upsertErr
}

func (m *mockSlotStore) SlotHasActiveBooking(ctx context.Context, slotID string) (bool, error) {
	return m.hasBooking, m.bookingErr
}

func TestGetAvailableSlots_RoomNotFound(t *testing.T) {
	roomStore := &mockRoomStore{roomExists: false}
	scheduleStore := &mockScheduleStore{}
	slotStore := &mockSlotStore{}

	svc := NewSlotService(roomStore, scheduleStore, slotStore)
	slots, err := svc.GetAvailableSlots(context.Background(), "nonexistent-room", time.Now())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected empty slots, got %d", len(slots))
	}
}

func TestGetAvailableSlots_NilSchedule(t *testing.T) {
	roomStore := &mockRoomStore{roomExists: true}
	scheduleStore := &mockScheduleStore{schedule: nil}
	slotStore := &mockSlotStore{}

	svc := NewSlotService(roomStore, scheduleStore, slotStore)
	slots, err := svc.GetAvailableSlots(context.Background(), "room-1", time.Now())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected empty slots for nil schedule, got %d", len(slots))
	}
}

func TestGetAvailableSlots_DayNotAllowed(t *testing.T) {
	roomStore := &mockRoomStore{roomExists: true}
	schedule := &model.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	scheduleStore := &mockScheduleStore{schedule: schedule}
	slotStore := &mockSlotStore{}

	svc := NewSlotService(roomStore, scheduleStore, slotStore)
	// time.Date(2026, 4, 4) is a Saturday (weekday 6)
	saturday := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	slots, err := svc.GetAvailableSlots(context.Background(), "room-1", saturday)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected no slots on Saturday, got %d", len(slots))
	}
}

func TestGetAvailableSlots_Success(t *testing.T) {
	roomStore := &mockRoomStore{roomExists: true}
	schedule := &model.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}
	scheduleStore := &mockScheduleStore{schedule: schedule}
	slotStore := &mockSlotStore{hasBooking: false}

	svc := NewSlotService(roomStore, scheduleStore, slotStore)
	// time.Date(2026, 4, 1) is a Wednesday (weekday 3)
	wednesday := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	slots, err := svc.GetAvailableSlots(context.Background(), "room-1", wednesday)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 2 {
		t.Errorf("expected 2 slots (09:00-09:30, 09:30-10:00), got %d", len(slots))
	}
}

func TestGetAvailableSlots_AllSlotsBooked(t *testing.T) {
	roomStore := &mockRoomStore{roomExists: true}
	schedule := &model.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}
	scheduleStore := &mockScheduleStore{schedule: schedule}
	slotStore := &mockSlotStore{hasBooking: true}

	svc := NewSlotService(roomStore, scheduleStore, slotStore)
	wednesday := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	slots, err := svc.GetAvailableSlots(context.Background(), "room-1", wednesday)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected 0 available slots when all booked, got %d", len(slots))
	}
}

func TestIsDayAllowed(t *testing.T) {
	allowed := []int{1, 2, 3, 4, 5}

	if !IsDayAllowed(1, allowed) {
		t.Error("Monday should be allowed")
	}
	if !IsDayAllowed(5, allowed) {
		t.Error("Friday should be allowed")
	}
	if IsDayAllowed(6, allowed) {
		t.Error("Saturday should not be allowed")
	}
	if IsDayAllowed(7, allowed) {
		t.Error("Sunday should not be allowed")
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		wantH int
		wantM int
	}{
		{"09:00", 9, 0},
		{"14:30", 14, 30},
		{"00:00", 0, 0},
		{"23:59", 23, 59},
		{"08:00", 8, 0},
	}
	for _, tt := range tests {
		h, m := ParseTime(tt.input)
		if h != tt.wantH || m != tt.wantM {
			t.Errorf("ParseTime(%s) = %d:%d, want %d:%d", tt.input, h, m, tt.wantH, tt.wantM)
		}
	}
}

func TestDeterministicSlotID(t *testing.T) {
	roomID := "room-1"
	start := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	id1 := DeterministicSlotID(roomID, start)
	id2 := DeterministicSlotID(roomID, start)

	if id1 != id2 {
		t.Errorf("slot IDs differ: %s vs %s", id1, id2)
	}

	differentStart := time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC)
	id3 := DeterministicSlotID(roomID, differentStart)
	if id1 == id3 {
		t.Error("different start times should produce different IDs")
	}
}

func TestIsSlotInPast(t *testing.T) {
	past := time.Now().UTC().Add(-1 * time.Hour)
	future := time.Now().UTC().Add(1 * time.Hour)

	if !IsSlotInPast(past) {
		t.Error("past slot should be in past")
	}
	if IsSlotInPast(future) {
		t.Error("future slot should not be in past")
	}
}

func TestValidateSchedule(t *testing.T) {
	tests := []struct {
		name       string
		days       []int
		start      string
		end        string
		wantStatus int
		wantMsg    string
	}{
		{"valid", []int{1, 2, 3}, "09:00", "18:00", 0, ""},
		{"empty days", []int{}, "09:00", "18:00", 400, "daysOfWeek must not be empty"},
		{"invalid day", []int{0}, "09:00", "18:00", 400, "daysOfWeek values must be 1-7"},
		{"day 8", []int{8}, "09:00", "18:00", 400, "daysOfWeek values must be 1-7"},
		{"start after end", []int{1}, "18:00", "09:00", 400, "startTime must be before endTime"},
		{"start equal end", []int{1}, "09:00", "09:00", 400, "startTime must be before endTime"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := ValidateSchedule(tt.days, tt.start, tt.end)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if msg != tt.wantMsg {
				t.Errorf("msg = %s, want %s", msg, tt.wantMsg)
			}
		})
	}
}

func TestNewSlotService(t *testing.T) {
	s := NewSlotService(nil, nil, nil)
	if s == nil {
		t.Error("NewSlotService should not return nil")
	}
}
