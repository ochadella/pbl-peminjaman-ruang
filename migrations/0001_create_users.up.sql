CREATE TABLE IF NOT EXISTS users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nim_nip       VARCHAR(20)  NOT NULL,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(100) NOT NULL,
    password_hash TEXT         NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa',
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    blocked_until TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT users_nim_nip_key UNIQUE (nim_nip),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_role_check CHECK (role IN ('mahasiswa', 'petugas', 'admin'))
);