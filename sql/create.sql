
-- For the remote / local machines
CREATE TABLE log
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    log_time TIMESTAMPTZ NOT NULL,
    message JSONB NOT NULL,
    context JSONB NOT NULL
);