-- +goose Up
-- +goose StatementBegin
CREATE TYPE orderStatus as ENUM ('REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    userID text,
    status orderStatus,
    accrual float,
    uploaded bigint
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TYPE IF EXISTS orderStatus CASCADE;

DROP TABLE IF EXISTS orders CASCADE;

SELECT 'down SQL query';
-- +goose StatementEnd
