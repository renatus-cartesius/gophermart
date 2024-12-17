-- +goose Up
-- +goose StatementBegin

CREATE TABLE withdrawals (
    orderID bigint,
    userID text,
    amount float,
    time bigint
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS withdrawals CASCADE;

SELECT 'down SQL query';

-- +goose StatementEnd