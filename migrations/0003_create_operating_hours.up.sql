CREATE TABLE IF NOT EXISTS operating_hours (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    day_of_week SMALLINT NOT NULL,
    open_time   TIME     NOT NULL,
    close_time  TIME     NOT NULL,
    is_open     BOOLEAN  NOT NULL DEFAULT TRUE,
    CONSTRAINT operating_hours_day_key UNIQUE (day_of_week),
    CONSTRAINT operating_hours_day_check CHECK (day_of_week BETWEEN 0 AND 6),
    CONSTRAINT operating_hours_time_check CHECK (NOT is_open OR close_time > open_time)
);