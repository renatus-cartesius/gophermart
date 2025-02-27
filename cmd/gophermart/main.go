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
	"github.com/renatus-cartesius/gophermart/internal/storage"
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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal(
			"error on loading config",
			zap.Error(err),
		)
	}

	a := accrual.NewAccrual(cfg.AccrualAddres)

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

	pgStorage := storage.NewPGStorage(db)

	l := loyalty.NewLoyalty(
		a,
		pgStorage,
	)

	srv := handlers.NewServerHandler(
		l,
		auth.NewAuth(
			[]byte("d6b32087c4b1f7c8b88c945234d54cfa5aa73d4b14e5e7a778448d515db00028b20db"),
			pgStorage,
		),
	)

	r := chi.NewRouter()
	server := &http.Server{Addr: cfg.SrvAddress, Handler: r}

	handlers.Setup(r, srv)

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	dispatchContext, dispatchContextCancel := context.WithCancel(context.Background())
	defer dispatchContextCancel()
	go l.Dispatch(dispatchContext)

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
