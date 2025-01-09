-- +goose Up
-- +goose StatementBegin
CREATE TYPE orderStatus as ENUM ('NEW', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE orders (
    id text PRIMARY KEY,
    userID text,
    status orderStatus,
    accrual float default 0,
    uploaded timestamp default (timezone('utc', now()))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS orderStatus;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
