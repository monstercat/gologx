CREATE TABLE log
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id UUID REFERENCES service (id),
    log_type   TEXT        NOT NULL,
    log_time   TIMESTAMPTZ NOT NULL,
    message    JSONB       NOT NULL,
    context    JSONB       NOT NULL
);

CREATE TABLE route_log
(
    id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    log_id   UUID REFERENCES log (id),
    method   TEXT  NOT NULL,
    severity TEXT,
    path     TEXT  NOT NULL,
    ip       TEXT  NOT NULL,
    body     JSONB NOT NULL,
    headers  JSONB NOT NULL
);

CREATE TABLE origin
(
    id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE service
(
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    origin_id UUID REFERENCES origin (id),
    name      TEXT NOT NULL UNIQUE,
    last_seen TIMESTAMPTZ,
    sig_hash  TEXT NOT NULL    DEFAULT ''
);