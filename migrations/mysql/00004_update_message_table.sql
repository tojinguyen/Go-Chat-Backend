-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
DROP COLUMN receiver_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages
ADD COLUMN receiver_id VARCHAR(36) NOT NULL;
-- +goose StatementEnd