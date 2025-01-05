package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/renatus-cartesius/gophermart/cmd/gophermart/config"
	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/internal/auth"
	"github.com/renatus-cartesius/gophermart/internal/loyalty"
	"github.com/renatus-cartesius/gophermart/internal/server/handlers"
	"github.com/renatus-cartesius/gophermart/pkg/logger"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	// ctx := context.Background()

	if err := logger.Initialize("DEBUG"); err != nil {
		log.Fatalln(err)
	}

	mockAccrualler := loyalty.MockAccrualler{
		Orders: map[int64]*accrual.OrderInfo{
			79927398713: {
				Order:   79927398713,
				Accrual: 331.3,
				Status:  accrual.TypeStatusProcessing,
			},
			3938230889: {
				Order:   3938230889,
				Accrual: 0,
				Status:  accrual.TypeStatusInvalid,
			},
			4929972884676289: {
				Order:   4929972884676289,
				Accrual: 999999,
				Status:  accrual.TypeStatusProcessed,
			},
			2050898812: {
				Order:   2050898812,
				Accrual: 43,
				Status:  accrual.TypeStatusRegistered,
			},
		},
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal(
			"error on loading config",
			zap.Error(err),
		)
	}

	db, err := sql.Open("pgx", cfg.DBURI)
	if err != nil {
		logger.Log.Fatal(
			"error on openning DB connection",
			zap.Error(err),
		)
	}
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Log.Fatal(
			"error on checking DB connection",
			zap.Error(pingErr),
		)
	}
	logger.Log.Debug(
		"successfully connected to pg db",
	)

	if err := migrationsHelper(db); err != nil {
		logger.Log.Fatal(
			"error on preparing or making migrations",
			zap.Error(err),
		)
	}

	loyaltyStorage := loyalty.NewLoyaltyStoragePg(db)

	srv := handlers.NewServerHandler(
		loyalty.NewLoyalty(
			mockAccrualler,
			loyaltyStorage,
		),
		auth.NewAuth(
			[]byte("d6b32087c4b1f7c8b88c945234d54cfa5aa73d4b14e5e7a778448d515db00028b20db"),
			db,
		),
	)

	r := chi.NewRouter()
	server := &http.Server{Addr: cfg.SrvAddress, Handler: r}

	handlers.Setup(r, srv)

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-shutdownSig

		logger.Log.Info(
			"graceful shuting down",
			zap.String("address", cfg.SrvAddress),
		)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Log.Fatal(
				"error on graceful shutdown",
				zap.String("address", cfg.SrvAddress),
			)
		}

	}()

	logger.Log.Info(
		"starting server",
		zap.String("address", cfg.SrvAddress),
	)

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}

}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func migrationsHelper(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	return errors.Join(goose.SetDialect("postgres"), goose.Up(db, "migrations"))
}
