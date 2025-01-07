-- +goose Up
-- +goose StatementBegin

CREATE TABLE withdrawals (
    orderID text,
    userID text,
    sum float,
    created timestamp default (timezone('utc', now()))
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd