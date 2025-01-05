-- +goose Up
-- +goose StatementBegin
CREATE TYPE orderStatus as ENUM ('REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    userID text,
    status orderStatus,
    accrual float,
    uploaded timestamp default (timezone('utc', now()))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS orderStatus;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
