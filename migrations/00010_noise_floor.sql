-- +goose Up
-- +goose StatementBegin
ALTER TABLE media_configuration
    ADD COLUMN noise_floor_db NUMERIC(5,2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE media_configuration
    DROP COLUMN IF EXISTS noise_floor_db;
-- +goose StatementEnd
