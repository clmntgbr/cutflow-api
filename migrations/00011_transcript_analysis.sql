-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE word_kind ADD VALUE IF NOT EXISTS 'false_start';

CREATE TYPE transcript_issue_type AS ENUM ('filler', 'repetition', 'false_start');

CREATE TABLE detected_transcript_issue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    transcript_id UUID NOT NULL REFERENCES transcript(id) ON DELETE CASCADE,
    issue_type transcript_issue_type NOT NULL,
    text TEXT NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    confidence NUMERIC(5,4),
    word_start_index INTEGER NOT NULL,
    word_end_index INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (word_start_index >= 0),
    CHECK (word_end_index >= word_start_index)
);

CREATE INDEX idx_detected_transcript_issue_media
    ON detected_transcript_issue(media_file_id, source_start_ms);

CREATE INDEX idx_detected_transcript_issue_transcript
    ON detected_transcript_issue(transcript_id, issue_type);

-- +goose Down
-- +goose NO TRANSACTION
DROP TABLE IF EXISTS detected_transcript_issue;
DROP TYPE IF EXISTS transcript_issue_type;
-- Note: PostgreSQL cannot easily remove an enum value; false_start remains on word_kind.
