-- +goose Up
-- +goose StatementBegin
CREATE TABLE timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    fingerprint VARCHAR(64) NOT NULL,
    engine_version VARCHAR(64) NOT NULL DEFAULT 'timeline-engine-v1',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (media_file_id, version),
    UNIQUE (media_file_id, fingerprint)
);

CREATE INDEX idx_timeline_media_active
    ON timeline(media_file_id, is_active)
    WHERE is_active = TRUE;

CREATE INDEX idx_timeline_project
    ON timeline(project_id);

CREATE TABLE timeline_segment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timeline_id UUID NOT NULL REFERENCES timeline(id) ON DELETE CASCADE,
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    output_start_ms BIGINT NOT NULL,
    output_end_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (output_start_ms >= 0),
    CHECK (output_end_ms > output_start_ms),
    UNIQUE (timeline_id, segment_index)
);

CREATE INDEX idx_timeline_segment_timeline
    ON timeline_segment(timeline_id, segment_index);

CREATE TABLE edit_decision (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    timeline_id UUID REFERENCES timeline(id) ON DELETE CASCADE,
    decision_type VARCHAR(32) NOT NULL,
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    action VARCHAR(16) NOT NULL,
    source VARCHAR(16) NOT NULL DEFAULT 'automatic',
    confidence NUMERIC(5,4),
    reasons JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (action IN ('remove', 'keep')),
    CHECK (source IN ('automatic', 'user'))
);

CREATE INDEX idx_edit_decision_media
    ON edit_decision(media_file_id, source_start_ms);

CREATE INDEX idx_edit_decision_timeline
    ON edit_decision(timeline_id);

CREATE TABLE user_override (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    override_type VARCHAR(32) NOT NULL DEFAULT 'manual',
    source_start_ms BIGINT NOT NULL,
    source_end_ms BIGINT NOT NULL,
    action VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source_start_ms >= 0),
    CHECK (source_end_ms > source_start_ms),
    CHECK (action IN ('remove', 'keep'))
);

CREATE INDEX idx_user_override_media
    ON user_override(media_file_id, source_start_ms);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_override;
DROP TABLE IF EXISTS edit_decision;
DROP TABLE IF EXISTS timeline_segment;
DROP TABLE IF EXISTS timeline;
-- +goose StatementEnd
