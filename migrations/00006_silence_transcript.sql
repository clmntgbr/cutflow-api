-- +goose Up
-- +goose StatementBegin
CREATE TYPE word_kind AS ENUM ('speech', 'filler', 'repetition');

CREATE TABLE detected_silence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    start_ms BIGINT NOT NULL,
    end_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_ms >= 0),
    CHECK (end_ms > start_ms)
);

CREATE INDEX idx_detected_silence_media_time
    ON detected_silence(media_file_id, start_ms);

CREATE TABLE transcript (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL UNIQUE REFERENCES media_file(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10),
    text TEXT,
    srt_storage_key TEXT NOT NULL DEFAULT '',
    ass_storage_key TEXT NOT NULL DEFAULT '',
    provider_job_id VARCHAR(100),
    status job_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transcript_status ON transcript(status);

CREATE TABLE transcript_word (
    id BIGSERIAL PRIMARY KEY,
    transcript_id UUID NOT NULL REFERENCES transcript(id) ON DELETE CASCADE,
    word_index INTEGER NOT NULL,
    text TEXT NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    confidence NUMERIC(5,4),
    kind word_kind NOT NULL DEFAULT 'speech',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms >= source_start_ms),
    CHECK (confidence IS NULL OR confidence BETWEEN 0 AND 1),
    UNIQUE (transcript_id, word_index)
);

CREATE INDEX idx_transcript_word_time
    ON transcript_word(transcript_id, source_start_ms);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transcript_word;
DROP TABLE IF EXISTS transcript;
DROP TABLE IF EXISTS detected_silence;
DROP TYPE IF EXISTS word_kind;
-- +goose StatementEnd
