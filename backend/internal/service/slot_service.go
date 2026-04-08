package service

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
)

type RoomStore interface {
	RoomExists(ctx context.Context, roomID string) (bool, error)
}

type ScheduleStore interface {
	GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error)
}

type SlotStore interface {
	UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error)
	SlotHasActiveBooking(ctx context.Context, slotID string) (bool, error)
}

type SlotService struct {
	rooms     RoomStore
	schedules ScheduleStore
	slots     SlotStore
}

func NewSlotService(rooms RoomStore, schedules ScheduleStore, slots SlotStore) *SlotService {
	return &SlotService{rooms: rooms, schedules: schedules, slots: slots}
}

func (s *SlotService) GetAvailableSlots(ctx context.Context, roomID string, date time.Time) ([]model.Slot, error) {
	exists, err := s.rooms.RoomExists(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []model.Slot{}, nil
	}

	sch, err := s.schedules.GetSchedule(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if sch == nil {
		return []model.Slot{}, nil
	}

	dayOfWeek := int(date.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	if !IsDayAllowed(dayOfWeek, sch.DaysOfWeek) {
		return []model.Slot{}, nil
	}

	slots, err := GenerateSlots(ctx, s.slots, roomID, date, sch)
	if err != nil {
		return nil, err
	}

	var available []model.Slot
	for _, sl := range slots {
		booked, err := s.slots.SlotHasActiveBooking(ctx, sl.ID)
		if err != nil {
			return nil, err
		}
		if !booked {
			available = append(available, sl)
		}
	}

	if available == nil {
		available = []model.Slot{}
	}
	return available, nil
}

func IsDayAllowed(dayOfWeek int, allowedDays []int) bool {
	for _, d := range allowedDays {
		if d == dayOfWeek {
			return true
		}
	}
	return false
}

func GenerateSlots(ctx context.Context, store SlotStore, roomID string, date time.Time, sch *model.Schedule) ([]model.Slot, error) {
	sh, sm := ParseTime(sch.StartTime)
	eh, em := ParseTime(sch.EndTime)

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

	var slots []model.Slot
	for t := dayStart; t.Add(30*time.Minute).Before(dayEnd) || t.Add(30*time.Minute).Equal(dayEnd); t = t.Add(30 * time.Minute) {
		slotEnd := t.Add(30 * time.Minute)
		sl, err := store.UpsertSlot(ctx, roomID, t, slotEnd)
		if err != nil {
			return nil, err
		}
		slots = append(slots, sl)
	}
	return slots, nil
}

func ParseTime(s string) (int, int) {
	var h, m int
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			for j := 0; j < i; j++ {
				h = h*10 + int(s[j]-'0')
			}
			for j := i + 1; j < len(s); j++ {
				m = m*10 + int(s[j]-'0')
			}
			break
		}
	}
	return h, m
}

func DeterministicSlotID(roomID string, start time.Time) string {
	h := sha256.Sum256([]byte(roomID + start.UTC().Format(time.RFC3339)))
	return uuid.Must(uuid.FromBytes(h[:16])).String()
}

func IsSlotInPast(slotStart time.Time) bool {
	return slotStart.Before(time.Now().UTC())
}

func ValidateSchedule(daysOfWeek []int, startTime, endTime string) (int, string) {
	if len(daysOfWeek) == 0 {
		return 400, "daysOfWeek must not be empty"
	}
	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return 400, "daysOfWeek values must be 1-7"
		}
	}
	if startTime >= endTime {
		return 400, "startTime must be before endTime"
	}
	return 0, ""
}
