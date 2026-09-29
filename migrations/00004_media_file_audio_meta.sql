-- +goose Up
-- +goose StatementBegin
ALTER TABLE media_file
    ADD COLUMN IF NOT EXISTS audio_sample_rate INTEGER,
    ADD COLUMN IF NOT EXISTS audio_channels INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE media_file
    DROP COLUMN IF EXISTS audio_channels,
    DROP COLUMN IF EXISTS audio_sample_rate;
-- +goose StatementEnd
