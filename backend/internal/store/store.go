package store

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/avito-test/room-booking/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func deterministicSlotID(roomID string, start time.Time) string {
	h := sha256.Sum256([]byte(roomID + start.UTC().Format(time.RFC3339)))
	return uuid.Must(uuid.FromBytes(h[:16])).String()
}

func (s *Store) CreateRoom(ctx context.Context, id, name string, desc *string, cap *int) (model.Room, error) {
	var r model.Room
	err := s.pool.QueryRow(ctx,
		`INSERT INTO rooms (id, name, description, capacity) VALUES ($1,$2,$3,$4)
		 RETURNING id::text, name, description, capacity, created_at::text`,
		id, name, desc, cap,
	).Scan(&r.ID, &r.Name, &r.Description, &r.Capacity, &r.CreatedAt)
	return r, err
}

func (s *Store) ListRooms(ctx context.Context) ([]model.Room, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, name, description, capacity, created_at::text FROM rooms ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []model.Room
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Capacity, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (s *Store) RoomExists(ctx context.Context, roomID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM rooms WHERE id=$1)`, roomID).Scan(&exists)
	return exists, err
}

func (s *Store) CreateSchedule(ctx context.Context, id, roomID string, days []int, start, end string) (model.Schedule, error) {
	var sch model.Schedule
	err := s.pool.QueryRow(ctx,
		`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
		 VALUES ($1,$2,$3,$4::time,$5::time)
		 RETURNING id::text, room_id::text, days_of_week, start_time::text, end_time::text`,
		id, roomID, days, start, end,
	).Scan(&sch.ID, &sch.RoomID, &sch.DaysOfWeek, &sch.StartTime, &sch.EndTime)
	return sch, err
}

func (s *Store) ScheduleExists(ctx context.Context, roomID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id=$1)`, roomID).Scan(&exists)
	return exists, err
}

func (s *Store) GetSchedule(ctx context.Context, roomID string) (*model.Schedule, error) {
	var sch model.Schedule
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, room_id::text, days_of_week, start_time::text, end_time::text
		 FROM schedules WHERE room_id=$1`, roomID,
	).Scan(&sch.ID, &sch.RoomID, &sch.DaysOfWeek, &sch.StartTime, &sch.EndTime)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sch, nil
}

func (s *Store) UpsertSlot(ctx context.Context, roomID string, start, end time.Time) (model.Slot, error) {
	id := deterministicSlotID(roomID, start)
	var sl model.Slot
	err := s.pool.QueryRow(ctx,
		`INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (room_id, start_time) DO UPDATE SET start_time=EXCLUDED.start_time
		 RETURNING id::text, room_id::text, start_time::text, end_time::text`,
		id, roomID, start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339),
	).Scan(&sl.ID, &sl.RoomID, &sl.Start, &sl.End)
	return sl, err
}

func (s *Store) GetSlot(ctx context.Context, slotID string) (*model.Slot, error) {
	var sl model.Slot
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, room_id::text, start_time::text, end_time::text
		 FROM slots WHERE id=$1`, slotID,
	).Scan(&sl.ID, &sl.RoomID, &sl.Start, &sl.End)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

func (s *Store) SlotHasActiveBooking(ctx context.Context, slotID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id=$1 AND status='active')`, slotID,
	).Scan(&exists)
	return exists, err
}

func (s *Store) CreateBooking(ctx context.Context, id, slotID, userID string, confLink *string) (model.Booking, error) {
	var b model.Booking
	err := s.pool.QueryRow(ctx,
		`INSERT INTO bookings (id, slot_id, user_id, conference_link) VALUES ($1,$2,$3,$4)
		 RETURNING id::text, slot_id::text, user_id::text, status, conference_link, created_at::text`,
		id, slotID, userID, confLink,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	return b, err
}

func (s *Store) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	var b model.Booking
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, slot_id::text, user_id::text, status, conference_link, created_at::text
		 FROM bookings WHERE id=$1`, id,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) CancelBooking(ctx context.Context, id string) (model.Booking, error) {
	var b model.Booking
	err := s.pool.QueryRow(ctx,
		`UPDATE bookings SET status='cancelled' WHERE id=$1
		 RETURNING id::text, slot_id::text, user_id::text, status, conference_link, created_at::text`,
		id,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	return b, err
}

func (s *Store) GetAllSlotsWithBookings(ctx context.Context, roomID string, dateStart, dateEnd time.Time) ([]model.SlotWithBooking, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT s.id::text, s.room_id::text, s.start_time::text, s.end_time::text,
		        CASE WHEN b.id IS NOT NULL THEN 'booked' ELSE 'free' END as status,
		        b.user_id::text, u.email
		 FROM slots s
		 LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		 LEFT JOIN users u ON u.id = b.user_id
		 WHERE s.room_id = $1 AND s.start_time >= $2 AND s.start_time < $3
		 ORDER BY s.start_time`,
		roomID, dateStart.UTC(), dateEnd.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []model.SlotWithBooking
	for rows.Next() {
		var sl model.SlotWithBooking
		var userID *string
		var userEmail *string
		if err := rows.Scan(&sl.ID, &sl.RoomID, &sl.Start, &sl.End, &sl.Status, &userID, &userEmail); err != nil {
			return nil, err
		}
		if userID != nil {
			sl.Booking = &model.SlotBookingInfo{UserID: *userID, UserEmail: *userEmail}
		}
		slots = append(slots, sl)
	}
	if slots == nil {
		slots = []model.SlotWithBooking{}
	}
	return slots, nil
}

func (s *Store) ListUserBookings(ctx context.Context, userID string) ([]model.Booking, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT b.id::text, b.slot_id::text, b.user_id::text, b.status, b.conference_link, b.created_at::text,
		        s.start_time::text, s.end_time::text, s.room_id::text, r.name
		 FROM bookings b
		 JOIN slots s ON s.id = b.slot_id
		 JOIN rooms r ON r.id = s.room_id
		 WHERE b.user_id = $1 AND s.start_time > now()
		 ORDER BY s.start_time`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
			&b.SlotStart, &b.SlotEnd, &b.RoomID, &b.RoomName); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (s *Store) ListAllBookings(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM bookings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx,
		`SELECT b.id::text, b.slot_id::text, b.user_id::text, b.status, b.conference_link, b.created_at::text,
		        u.email, s.start_time::text, s.end_time::text, r.name
		 FROM bookings b
		 JOIN slots s ON s.id = b.slot_id
		 JOIN rooms r ON r.id = s.room_id
		 JOIN users u ON u.id = b.user_id
		 ORDER BY b.created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
			&b.UserEmail, &b.SlotStart, &b.SlotEnd, &b.RoomName); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	return bookings, total, nil
}

func (s *Store) EnsureUserExists(ctx context.Context, id, email, role string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, email, role) VALUES ($1,$2,$3) ON CONFLICT (id) DO NOTHING`,
		id, email, role)
	return err
}

func (s *Store) CreateUser(ctx context.Context, id, email, password, role string) (model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (id, email, password, role) VALUES ($1,$2,$3,$4)
		 RETURNING id::text, email, role, created_at::text`,
		id, email, password, role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt)
	return u, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, email, password, role, created_at::text FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
