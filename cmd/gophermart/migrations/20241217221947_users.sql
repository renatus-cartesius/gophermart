-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id text PRIMARY KEY,
    passwordHash text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
