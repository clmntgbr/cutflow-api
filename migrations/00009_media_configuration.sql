-- +goose Up
-- +goose StatementBegin
CREATE TYPE silence_threshold_mode AS ENUM ('auto', 'manual');

CREATE TYPE silence_detection_level AS ENUM (
    'low',
    'moderate',
    'aggressive',
    'very_aggressive'
);

CREATE TABLE media_configuration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_file_id UUID NOT NULL UNIQUE REFERENCES media_file(id) ON DELETE CASCADE,

    silence_removal_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    silence_threshold_mode silence_threshold_mode NOT NULL DEFAULT 'auto',
    silence_threshold_db NUMERIC(5,2),
    calculated_silence_threshold_db NUMERIC(5,2),
    silence_detection_level silence_detection_level NOT NULL DEFAULT 'aggressive',
    silence_padding_before_ms INTEGER NOT NULL DEFAULT 50,
    silence_padding_after_ms INTEGER NOT NULL DEFAULT 150,
    silence_min_duration_ms INTEGER NOT NULL DEFAULT 500,
    speech_min_duration_ms INTEGER NOT NULL DEFAULT 300,

    filler_removal_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    repetition_removal_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    subtitles_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    subtitle_max_words SMALLINT NOT NULL DEFAULT 3,

    viral_detection_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    viral_clip_min_duration_ms INTEGER NOT NULL DEFAULT 20000,
    viral_clip_max_duration_ms INTEGER NOT NULL DEFAULT 90000,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (silence_padding_before_ms >= 0),
    CHECK (silence_padding_after_ms >= 0),
    CHECK (silence_min_duration_ms >= 0),
    CHECK (speech_min_duration_ms >= 0),
    CHECK (
        silence_threshold_mode <> 'manual'
        OR silence_threshold_db IS NOT NULL
    ),
    CHECK (subtitle_max_words BETWEEN 1 AND 10),
    CHECK (
        viral_clip_min_duration_ms > 0
        AND viral_clip_max_duration_ms >= viral_clip_min_duration_ms
    )
);

CREATE INDEX idx_media_configuration_media ON media_configuration(media_file_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_configuration;
DROP TYPE IF EXISTS silence_detection_level;
DROP TYPE IF EXISTS silence_threshold_mode;
-- +goose StatementEnd
