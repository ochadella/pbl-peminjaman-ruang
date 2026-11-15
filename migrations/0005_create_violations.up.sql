CREATE TABLE IF NOT EXISTS violations (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id),
    booking_id  BIGINT      REFERENCES bookings (id),
    type        VARCHAR(30) NOT NULL,
    note        TEXT,
    recorded_by BIGINT      REFERENCES users (id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT violations_booking_id_key UNIQUE (booking_id)
);

CREATE INDEX IF NOT EXISTS idx_violations_user ON violations (user_id);