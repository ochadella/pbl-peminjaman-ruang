CREATE TABLE IF NOT EXISTS rooms (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code       VARCHAR(20)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    location   VARCHAR(100) NOT NULL,
    capacity   INTEGER      NOT NULL,
    facilities TEXT,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT rooms_code_key UNIQUE (code),
    CONSTRAINT rooms_capacity_check CHECK (capacity > 0)
);