package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/internal/loyalty"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func (l *PGStorage) AddOrder(ctx context.Context, userID string, orderInfo *accrual.OrderInfo) error {

	// Check if order already in db by that user
	var uID string

	existingOrderRow := l.db.QueryRowContext(ctx, "SELECT userID FROM orders WHERE id = $1", orderInfo.Order)

	err := existingOrderRow.Err()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug(
			"error on executing query",
			zap.Error(err),
		)
	}

	if err := existingOrderRow.Scan(&uID); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			_, err = l.db.ExecContext(ctx, "INSERT INTO orders (id, userID, status, accrual) VALUES ($1, $2, $3, $4)", orderInfo.Order, userID, orderInfo.Status, orderInfo.Accrual)
			return err
		}

		logger.Log.Debug(
			"error on scanning into string",
			zap.Error(err),
		)
	}

	if uID != userID {
		return loyalty.ErrOrderUploadedAnotherUser
	} else {
		return loyalty.ErrOrderAlreadyUploaded
	}

}

func (l *PGStorage) GetOrders(ctx context.Context, userID string) ([]*loyalty.Order, error) {

	orders := make([]*loyalty.Order, 0)

	rows, err := l.db.QueryContext(ctx, "SELECT * FROM orders WHERE userID = $1 ORDER BY uploaded", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		order := &loyalty.Order{}
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.Uploaded); err != nil {
			logger.Log.Debug(
				"error on scanning row to Order",
				zap.Error(err),
			)
			continue
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (l *PGStorage) GetWithdrawals(ctx context.Context, userID string) ([]*loyalty.Withdraw, error) {

	withdrawals := make([]*loyalty.Withdraw, 0)

	rows, err := l.db.QueryContext(ctx, "SELECT * FROM withdrawals WHERE userID = $1 ORDER BY created", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		withdraw := &loyalty.Withdraw{}
		if err := rows.Scan(&withdraw.OrderID, &withdraw.UserID, &withdraw.Sum, &withdraw.Created); err != nil {
			logger.Log.Debug(
				"error on scanning row to Order",
				zap.Error(err),
			)
			continue
		}
		withdrawals = append(withdrawals, withdraw)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (l *PGStorage) GetOrder(ctx context.Context, orderID string) (*loyalty.Order, error) {
	return nil, nil
}

func (l *PGStorage) GetBalance(ctx context.Context, userID string) (*loyalty.Balance, error) {
	row := l.db.QueryRowContext(ctx, `
		select 
			((select COALESCE(SUM(accrual), 0) from orders where userID = $1) -
			(select COALESCE(SUM(sum), 0) from withdrawals where userID = $1)) as balance,
			(select COALESCE(SUM(sum), 0) from withdrawals where userID = $1) as withdrawn;
		;
	`, userID)

	balance := &loyalty.Balance{}
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		logger.Log.Debug(
			"error on scanning to row to balance",
		)
		return nil, err
	}

	return balance, row.Err()
}

func (l *PGStorage) AddWithdraw(ctx context.Context, wr *loyalty.Withdraw) error {
	_, err := l.db.ExecContext(ctx, "INSERT INTO withdrawals (orderID, userID, sum) VALUES ($1, $2, $3)", wr.OrderID, wr.UserID, wr.Sum)
	return err
}
