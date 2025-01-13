package storage

import (
	"context"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func (pg *PGStorage) IsUserExists(ctx context.Context, userID string) (bool, error) {

	row := pg.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT * FROM users WHERE id = $1)", userID)

	var userExists bool
	if err := row.Scan(&userExists); err != nil {
		logger.Log.Debug(
			"error on scanning row into bool",
			zap.Error(err),
		)
		return false, err
	}

	return userExists, row.Err()

}

func (pg *PGStorage) AddUser(ctx context.Context, userID, passwordHash string) error {
	_, err := pg.db.ExecContext(ctx, "INSERT INTO users (id, passwordHash) VALUES ($1, $2)", userID, passwordHash)
	return err
}

func (pg *PGStorage) GetHash(ctx context.Context, userID string) (string, error) {
	var realpasswordHash string
	hashRow := pg.db.QueryRowContext(ctx, "SELECT passwordHash from users where id = $1", userID)
	if err := hashRow.Scan(&realpasswordHash); err != nil {
		logger.Log.Debug(
			"error on scanning row into string",
			zap.Error(err),
		)
		return "", err
	}

	return realpasswordHash, hashRow.Err()
}
