package model

import "testing"

func TestRoomJSON(t *testing.T) {
	desc := "Test room"
	cap := 10
	r := Room{
		ID:          "room-1",
		Name:        "Room A",
		Description: &desc,
		Capacity:    &cap,
		CreatedAt:   "2026-04-01T09:00:00Z",
	}
	if r.ID != "room-1" {
		t.Errorf("ID = %s, want room-1", r.ID)
	}
	if r.Name != "Room A" {
		t.Errorf("Name = %s, want Room A", r.Name)
	}
	if *r.Description != "Test room" {
		t.Errorf("Description = %s, want Test room", *r.Description)
	}
	if *r.Capacity != 10 {
		t.Errorf("Capacity = %d, want 10", *r.Capacity)
	}
}

func TestScheduleJSON(t *testing.T) {
	s := Schedule{
		ID:         "sch-1",
		RoomID:     "room-1",
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	if s.RoomID != "room-1" {
		t.Errorf("RoomID = %s, want room-1", s.RoomID)
	}
	if len(s.DaysOfWeek) != 5 {
		t.Errorf("DaysOfWeek len = %d, want 5", len(s.DaysOfWeek))
	}
}

func TestSlotJSON(t *testing.T) {
	s := Slot{
		ID:     "slot-1",
		RoomID: "room-1",
		Start:  "2026-04-01T09:00:00Z",
		End:    "2026-04-01T09:30:00Z",
	}
	if s.Start != "2026-04-01T09:00:00Z" {
		t.Errorf("Start = %s, want 2026-04-01T09:00:00Z", s.Start)
	}
}

func TestBookingJSON(t *testing.T) {
	link := "https://meet.example.com/abc"
	b := Booking{
		ID:             "booking-1",
		SlotID:         "slot-1",
		UserID:         "user-1",
		Status:         "active",
		ConferenceLink: &link,
		CreatedAt:      "2026-04-01T09:00:00Z",
	}
	if b.Status != "active" {
		t.Errorf("Status = %s, want active", b.Status)
	}
	if *b.ConferenceLink != "https://meet.example.com/abc" {
		t.Errorf("ConferenceLink = %s, want https://meet.example.com/abc", *b.ConferenceLink)
	}
}

func TestBookingNilFields(t *testing.T) {
	b := Booking{
		ID:     "booking-1",
		SlotID: "slot-1",
		UserID: "user-1",
		Status: "cancelled",
	}
	if b.ConferenceLink != nil {
		t.Error("ConferenceLink should be nil")
	}
}
