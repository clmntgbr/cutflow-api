-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM (
            'pending', 'queued', 'processing', 'completed', 'failed', 'cancelled'
        );
    END IF;
END $$;

CREATE TABLE segment (
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

CREATE INDEX idx_segment_media ON segment(media_file_id);
CREATE INDEX idx_segment_status ON segment(status);

CREATE TABLE media_audio (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL UNIQUE REFERENCES media_file(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL,
    codec VARCHAR(50),
    sample_rate INTEGER,
    channels SMALLINT,
    size_bytes BIGINT,
    status job_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT media_audio_storage_key_key UNIQUE (storage_key)
);

CREATE INDEX idx_media_audio_project ON media_audio(project_id);
CREATE INDEX idx_media_audio_user ON media_audio(user_id);
CREATE INDEX idx_media_audio_status ON media_audio(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_audio;
DROP TABLE IF EXISTS segment;
DROP TYPE IF EXISTS job_status;
-- +goose StatementEnd
