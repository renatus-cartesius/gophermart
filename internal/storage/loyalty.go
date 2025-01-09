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

func (pg *PGStorage) AddOrder(ctx context.Context, userID string, orderID string) error {

	logger.Log.Debug(
		"begin adding order to storage",
		zap.String("orderID", orderID),
	)

	// Check if order already in db by that user
	var uID string

	existingOrderRow := pg.db.QueryRowContext(ctx, "SELECT userID FROM orders WHERE id = $1", orderID)

	err := existingOrderRow.Err()

	logger.Log.Debug(
		"checked order in storage",
		zap.String("orderID", orderID),
		zap.Error(err),
	)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug(
			"error on executing query",
			zap.Error(err),
		)
	}

	if err := existingOrderRow.Scan(&uID); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Debug(
				"inserting order to database",
				zap.String("orderID", orderID),
				zap.Error(err),
			)
			_, err = pg.db.ExecContext(ctx, "INSERT INTO orders (id, userID, status) VALUES ($1, $2, $3)", orderID, userID, loyalty.TypeStatusNew)
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

func (pg *PGStorage) GetOrders(ctx context.Context, userID string) ([]*loyalty.Order, error) {

	orders := make([]*loyalty.Order, 0)

	rows, err := pg.db.QueryContext(ctx, "SELECT * FROM orders WHERE userID = $1 ORDER BY uploaded", userID)
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

func (pg *PGStorage) GetWithdrawals(ctx context.Context, userID string) ([]*loyalty.Withdraw, error) {

	withdrawals := make([]*loyalty.Withdraw, 0)

	rows, err := pg.db.QueryContext(ctx, "SELECT * FROM withdrawals WHERE userID = $1 ORDER BY created", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		withdraw := &loyalty.Withdraw{}
		if err := rows.Scan(&withdraw.OrderID, &withdraw.UserID, &withdraw.Sum, &withdraw.Created); err != nil {
			logger.Log.Debug(
				"error on scanning row to Withdraw",
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

func (pg *PGStorage) GetOrder(ctx context.Context, orderID string) (*loyalty.Order, error) {
	orderRow := pg.db.QueryRowContext(ctx, "SELECT * FROM orders where id = $1", orderID)

	order := &loyalty.Order{}

	if err := orderRow.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.Uploaded); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, loyalty.ErrOrderNotFound
		}
		logger.Log.Debug(
			"error on scanning row to Order",
			zap.Error(err),
		)
		return nil, err
	}

	return order, orderRow.Err()

}

func (pg *PGStorage) GetBalance(ctx context.Context, userID string) (*loyalty.Balance, error) {
	row := pg.db.QueryRowContext(ctx, `
		select 
			((select COALESCE(SUM(accrual), 0) from orders where userID = $1 and status = $2) -
			(select COALESCE(SUM(sum), 0) from withdrawals where userID = $1)) as balance,
			(select COALESCE(SUM(sum), 0) from withdrawals where userID = $1) as withdrawn;
		;
	`, userID, loyalty.TypeStatusProcessed)

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

func (pg *PGStorage) AddWithdraw(ctx context.Context, wr *loyalty.Withdraw) error {
	_, err := pg.db.ExecContext(ctx, "INSERT INTO withdrawals (orderID, userID, sum) VALUES ($1, $2, $3)", wr.OrderID, wr.UserID, wr.Sum)
	return err
}

func (pg *PGStorage) GetUnhandledOrders(ctx context.Context) ([]string, error) {
	orders := make([]string, 0)

	rows, err := pg.db.QueryContext(ctx, "SELECT id from orders where status != $1 and status != $2", loyalty.TypeStatusProcessed, loyalty.TypeStatusInvalid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			logger.Log.Debug(
				"error on scanning row to string",
				zap.Error(err),
			)
			continue
		}
		orders = append(orders, orderID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
func (pg *PGStorage) UpdateOrder(ctx context.Context, orderInfo *accrual.OrderInfo) error {
	_, err := pg.db.ExecContext(ctx, "UPDATE orders SET status=$1, accrual=$2 WHERE id=$3", orderInfo.Status, orderInfo.Accrual, orderInfo.Order)
	return err
}
