CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE service
(
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    machine   TEXT NOT NULL UNIQUE,
    name      TEXT NOT NULL UNIQUE,
    last_seen TIMESTAMPTZ,
    sig_hash  TEXT NOT NULL    DEFAULT ''
);

CREATE TABLE log
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id UUID REFERENCES service (id),
    created    TIMESTAMPTZ      DEFAULT NOW(),
    log_type   TEXT        NOT NULL,
    log_time   TIMESTAMPTZ NOT NULL,
    message    TEXT        NOT NULL,
    context    JSONB       NOT NULL
);
