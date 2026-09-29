-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS segment;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM (
            'pending', 'queued', 'processing', 'completed', 'failed', 'cancelled'
        );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS segment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL,
    start_ms BIGINT NOT NULL,
    end_ms BIGINT NOT NULL,
    processing_start_ms BIGINT NOT NULL,
    processing_end_ms BIGINT NOT NULL,
    status job_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (segment_index >= 0),
    CHECK (start_ms >= 0),
    CHECK (end_ms > start_ms),
    CHECK (processing_start_ms >= 0),
    CHECK (processing_end_ms > processing_start_ms),
    UNIQUE (media_file_id, segment_index)
);

CREATE INDEX IF NOT EXISTS idx_segment_media ON segment(media_file_id);
CREATE INDEX IF NOT EXISTS idx_segment_status ON segment(status);
-- +goose StatementEnd
