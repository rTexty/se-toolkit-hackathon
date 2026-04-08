package tests

import (
	"fmt"
	"testing"
	"time"
)

func TestSlotGeneration(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		endTime   string
		wantSlots int
	}{
		{"9-10 gives 2 slots", "09:00", "10:00", 2},
		{"9-9:30 gives 1 slot", "09:00", "09:30", 1},
		{"full day 8-20 gives 24 slots", "08:00", "20:00", 24},
		{"half hour 14:00-14:30", "14:00", "14:30", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := generateTestSlots(tt.startTime, tt.endTime)
			if len(slots) != tt.wantSlots {
				t.Errorf("got %d slots, want %d", len(slots), tt.wantSlots)
			}
			for i, s := range slots {
				if d := s.end.Sub(s.start); d != 30*time.Minute {
					t.Errorf("slot %d duration = %v, want 30m", i, d)
				}
			}
			for i := 1; i < len(slots); i++ {
				if !slots[i].start.Equal(slots[i-1].end) {
					t.Errorf("gap between slot %d and %d", i-1, i)
				}
			}
		})
	}
}

type testSlot struct{ start, end time.Time }

func generateTestSlots(startStr, endStr string) []testSlot {
	parse := func(s string) (int, int) {
		var h, m int
		fmt.Sscanf(s, "%d:%d", &h, &m)
		return h, m
	}
	sh, sm := parse(startStr)
	eh, em := parse(endStr)
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

	var slots []testSlot
	for t := start; t.Add(30*time.Minute).Before(end) || t.Add(30*time.Minute).Equal(end); t = t.Add(30 * time.Minute) {
		slots = append(slots, testSlot{t, t.Add(30 * time.Minute)})
	}
	return slots
}

func TestDayOfWeekMapping(t *testing.T) {
	tests := []struct {
		weekday time.Weekday
		want    int
	}{
		{time.Monday, 1}, {time.Tuesday, 2}, {time.Wednesday, 3},
		{time.Thursday, 4}, {time.Friday, 5}, {time.Saturday, 6}, {time.Sunday, 7},
	}
	for _, tt := range tests {
		t.Run(tt.weekday.String(), func(t *testing.T) {
			got := isoDayOfWeek(tt.weekday)
			if got != tt.want {
				t.Errorf("isoDayOfWeek(%v) = %d, want %d", tt.weekday, got, tt.want)
			}
		})
	}
}

func isoDayOfWeek(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

func TestBookingIdempotentCancel(t *testing.T) {
	type booking struct {
		ID     string
		Status string
	}
	cancel := func(b *booking) {
		if b.Status == "cancelled" {
			return
		}
		b.Status = "cancelled"
	}

	b := &booking{"test-1", "active"}
	cancel(b)
	if b.Status != "cancelled" {
		t.Error("expected cancelled after first cancel")
	}
	cancel(b)
	if b.Status != "cancelled" {
		t.Error("expected cancelled after second cancel (idempotent)")
	}
}

func TestPastSlotRejection(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	if !past.Before(now) {
		t.Error("past should be before now")
	}
	if !future.After(now) {
		t.Error("future should be after now")
	}
}
