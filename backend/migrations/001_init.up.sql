CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL DEFAULT '',
    role       TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    capacity    INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS schedules (
    id           UUID PRIMARY KEY,
    room_id      UUID NOT NULL REFERENCES rooms(id),
    days_of_week INTEGER[] NOT NULL,
    start_time   TIME NOT NULL,
    end_time     TIME NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT one_schedule_per_room UNIQUE (room_id)
);

CREATE TABLE IF NOT EXISTS slots (
    id         UUID PRIMARY KEY,
    room_id    UUID NOT NULL REFERENCES rooms(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    CONSTRAINT unique_slot_per_room UNIQUE (room_id, start_time)
);

CREATE INDEX IF NOT EXISTS idx_slots_room_date ON slots (room_id, start_time);

CREATE TABLE IF NOT EXISTS bookings (
    id              UUID PRIMARY KEY,
    slot_id         UUID NOT NULL REFERENCES slots(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    conference_link TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT one_active_booking_per_slot
        EXCLUDE USING btree (slot_id WITH =) WHERE (status = 'active')
);

CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings (user_id, status);
CREATE INDEX IF NOT EXISTS idx_bookings_slot ON bookings (slot_id);
