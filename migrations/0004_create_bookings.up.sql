CREATE TABLE IF NOT EXISTS bookings (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES users (id),
    room_id           BIGINT       NOT NULL REFERENCES rooms (id),
    start_at          TIMESTAMPTZ  NOT NULL,
    end_at            TIMESTAMPTZ  NOT NULL,
    purpose           VARCHAR(200) NOT NULL,
    participant_count INTEGER      NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'booked',
    checkin_code      VARCHAR(12),
    checked_in_at     TIMESTAMPTZ,
    checked_out_at    TIMESTAMPTZ,
    cancelled_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT bookings_checkin_code_key UNIQUE (checkin_code),
    CONSTRAINT bookings_time_check CHECK (end_at > start_at),
    CONSTRAINT bookings_participant_check CHECK (participant_count > 0),
    CONSTRAINT bookings_status_check CHECK (status IN ('booked', 'checked_in', 'completed', 'cancelled', 'no_show'))
);

CREATE INDEX IF NOT EXISTS idx_bookings_room_time ON bookings (room_id, start_at, end_at);
CREATE INDEX IF NOT EXISTS idx_bookings_user_start ON bookings (user_id, start_at);