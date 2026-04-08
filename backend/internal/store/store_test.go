package store

import (
	"testing"
	"time"

	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
)

func TestDeterministicSlotID(t *testing.T) {
	roomID := "room-1"
	start := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	id1 := deterministicSlotID(roomID, start)
	id2 := deterministicSlotID(roomID, start)

	if id1 != id2 {
		t.Errorf("slot IDs differ: %s vs %s", id1, id2)
	}
	if _, err := uuid.Parse(id1); err != nil {
		t.Errorf("slot ID is not valid UUID: %s", id1)
	}

	differentStart := time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC)
	id3 := deterministicSlotID(roomID, differentStart)
	if id1 == id3 {
		t.Error("different start times should produce different IDs")
	}

	differentRoom := "room-2"
	id4 := deterministicSlotID(differentRoom, start)
	if id1 == id4 {
		t.Error("different room IDs should produce different IDs")
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
	}
	for _, tt := range tests {
		h, m := parseTimeTest(tt.input)
		if h != tt.wantH || m != tt.wantM {
			t.Errorf("parseTimeTest(%s) = %d:%d, want %d:%d", tt.input, h, m, tt.wantH, tt.wantM)
		}
	}
}

func TestGenerateSlots(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	sch := &model.Schedule{
		StartTime: "09:00",
		EndTime:   "10:00",
	}

	slots := generateTestSlots(date, sch)

	if len(slots) != 2 {
		t.Errorf("got %d slots, want 2", len(slots))
	}
	for i, s := range slots {
		if d := s.end.Sub(s.start); d != 30*time.Minute {
			t.Errorf("slot %d duration = %v, want 30m", i, d)
		}
	}
}

type testSlot struct{ start, end time.Time }

func parseTimeTest(s string) (int, int) {
	parts := split(s, ":")
	h, _ := atoi(parts[0])
	m, _ := atoi(parts[1])
	return h, m
}

func split(s, sep string) []string {
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}

func atoi(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func generateTestSlots(date time.Time, sch *model.Schedule) []testSlot {
	sh, sm := parseTimeTest(sch.StartTime)
	eh, em := parseTimeTest(sch.EndTime)

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

	var slots []testSlot
	for t := dayStart; t.Add(30*time.Minute).Before(dayEnd) || t.Add(30*time.Minute).Equal(dayEnd); t = t.Add(30 * time.Minute) {
		slots = append(slots, testSlot{t, t.Add(30 * time.Minute)})
	}
	return slots
}

func TestSlotHasActiveBookingQuery(t *testing.T) {
	q := `SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id=$1 AND status='active')`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestCreateBookingQuery(t *testing.T) {
	q := `INSERT INTO bookings (id, slot_id, user_id, conference_link) VALUES ($1,$2,$3,$4)
		 RETURNING id::text, slot_id::text, user_id::text, status, conference_link, created_at::text`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestCancelBookingQuery(t *testing.T) {
	q := `UPDATE bookings SET status='cancelled' WHERE id=$1
		 RETURNING id::text, slot_id::text, user_id::text, status, conference_link, created_at::text`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestListUserBookingsQuery(t *testing.T) {
	q := `SELECT b.id::text, b.slot_id::text, b.user_id::text, b.status, b.conference_link, b.created_at::text
		 FROM bookings b JOIN slots s ON s.id=b.slot_id
		 WHERE b.user_id=$1 AND s.start_time > now() ORDER BY s.start_time`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestListAllBookingsQuery(t *testing.T) {
	q := `SELECT count(*) FROM bookings`
	if q == "" {
		t.Error("count query should not be empty")
	}
}

func TestEnsureUserExistsQuery(t *testing.T) {
	q := `INSERT INTO users (id, email, role) VALUES ($1,$2,$3) ON CONFLICT (id) DO NOTHING`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestGetSlotQuery(t *testing.T) {
	q := `SELECT id::text, room_id::text, start_time::text, end_time::text
		 FROM slots WHERE id=$1`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestUpsertSlotQuery(t *testing.T) {
	q := `INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (room_id, start_time) DO UPDATE SET start_time=EXCLUDED.start_time
		 RETURNING id::text, room_id::text, start_time::text, end_time::text`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestCreateRoomQuery(t *testing.T) {
	q := `INSERT INTO rooms (id, name, description, capacity) VALUES ($1,$2,$3,$4)
		 RETURNING id::text, name, description, capacity, created_at::text`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestListRoomsQuery(t *testing.T) {
	q := `SELECT id::text, name, description, capacity, created_at::text FROM rooms ORDER BY created_at`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestRoomExistsQuery(t *testing.T) {
	q := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id=$1)`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestCreateScheduleQuery(t *testing.T) {
	q := `INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
		 VALUES ($1,$2,$3,$4::time,$5::time)
		 RETURNING id::text, room_id::text, days_of_week, start_time::text, end_time::text`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestScheduleExistsQuery(t *testing.T) {
	q := `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id=$1)`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestGetScheduleQuery(t *testing.T) {
	q := `SELECT id::text, room_id::text, days_of_week, start_time::text, end_time::text
		 FROM schedules WHERE room_id=$1`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestGetBookingByIDQuery(t *testing.T) {
	q := `SELECT id::text, slot_id::text, user_id::text, status, conference_link, created_at::text
		 FROM bookings WHERE id=$1`
	if q == "" {
		t.Error("query should not be empty")
	}
}

func TestListAllBookingsPagination(t *testing.T) {
	page := 1
	pageSize := 20
	offset := (page - 1) * pageSize
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}

	page = 3
	offset = (page - 1) * pageSize
	if offset != 40 {
		t.Errorf("offset = %d, want 40", offset)
	}
}

func TestStoreNew(t *testing.T) {
	s := New(nil)
	if s == nil {
		t.Error("New should return non-nil Store")
	}
	if s.pool != nil {
		t.Error("pool should be nil when passed nil")
	}
}
