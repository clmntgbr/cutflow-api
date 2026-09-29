-- +goose Up
-- +goose StatementBegin
CREATE TYPE task_status AS ENUM ('pending', 'processing', 'success', 'failed');

CREATE TABLE job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    media_file_id UUID NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    status task_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_job_project ON job(project_id);
CREATE INDEX idx_job_media_file ON job(media_file_id);
CREATE INDEX idx_job_status ON job(status);
CREATE INDEX idx_job_project_created ON job(project_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS job;
DROP TYPE IF EXISTS task_status;
-- +goose StatementEnd
