-- +goose Up
-- +goose StatementBegin
CREATE TYPE project_status AS ENUM (
    'draft', 'processing', 'ready', 'rendering', 'completed', 'failed'
);

CREATE TYPE media_status AS ENUM (
    'pending', 'uploaded', 'probing', 'ready', 'processing', 'completed', 'failed'
);

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status project_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_user ON project(user_id);
CREATE INDEX idx_project_status ON project(status);

CREATE TABLE media_file (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL,
    thumbnail_key TEXT NOT NULL DEFAULT '',
    original_filename TEXT,
    mime_type VARCHAR(100),
    duration_ms BIGINT NOT NULL DEFAULT 0,
    width INTEGER,
    height INTEGER,
    fps NUMERIC(8,3),
    video_codec VARCHAR(50),
    audio_codec VARCHAR(50),
    size_bytes BIGINT,
    status media_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT media_file_storage_key_key UNIQUE (storage_key),
    CHECK (duration_ms >= 0)
);

CREATE INDEX idx_media_file_project ON media_file(project_id);
CREATE INDEX idx_media_file_user ON media_file(user_id);
CREATE INDEX idx_media_file_status ON media_file(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_file;
DROP TABLE IF EXISTS project;
DROP TYPE IF EXISTS media_status;
DROP TYPE IF EXISTS project_status;
-- +goose StatementEnd
