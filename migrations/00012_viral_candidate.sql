-- +goose Up
-- +goose StatementBegin
ALTER TABLE media_configuration
    ADD COLUMN viral_max_candidates INTEGER NOT NULL DEFAULT 10,
    ADD COLUMN viral_min_score NUMERIC(5,4) NOT NULL DEFAULT 0.70;

ALTER TABLE media_configuration
    ADD CONSTRAINT media_configuration_viral_max_candidates_check
        CHECK (viral_max_candidates BETWEEN 1 AND 50),
    ADD CONSTRAINT media_configuration_viral_min_score_check
        CHECK (viral_min_score >= 0 AND viral_min_score <= 1);

CREATE TABLE viral_candidate (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    score NUMERIC(5,4) NOT NULL,
    hook_score NUMERIC(5,4),
    standalone_score NUMERIC(5,4),
    payoff_score NUMERIC(5,4),
    interest_score NUMERIC(5,4),
    title TEXT NOT NULL DEFAULT '',
    hook TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    provider VARCHAR(50) NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT '',
    selected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (score >= 0 AND score <= 1)
);

CREATE INDEX idx_viral_candidate_media_score
    ON viral_candidate(media_file_id, score DESC);

CREATE INDEX idx_viral_candidate_project
    ON viral_candidate(project_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS viral_candidate;

ALTER TABLE media_configuration
    DROP CONSTRAINT IF EXISTS media_configuration_viral_max_candidates_check,
    DROP CONSTRAINT IF EXISTS media_configuration_viral_min_score_check;

ALTER TABLE media_configuration
    DROP COLUMN IF EXISTS viral_max_candidates,
    DROP COLUMN IF EXISTS viral_min_score;
-- +goose StatementEnd
