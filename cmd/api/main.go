package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"final/internal/data"
	"final/internal/mailer"

	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	mailtrap    struct {
		apiKey string
		sender string
		url    string
	}
}

type applicationDependencies struct {
	config          serverConfig
	logger          *slog.Logger
	db              *sql.DB
	wg              sync.WaitGroup
	mailer          mailer.Mailer
	userModel       data.UserModel
	tokenModel      data.TokenModel
	permissionModel data.PermissionModel
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development",
		"Environment(development|staging|production)")

	flag.StringVar(&settings.mailtrap.url, "mailtrap-url", "https://sandbox.api.mailtrap.io/api/send/4520864", "Mailtrap URL")
	flag.StringVar(&settings.mailtrap.apiKey, "mailtrap-api-key", "9e9b8eb6387467d788719e69df04e061", "Mailtrap API Key")
	flag.StringVar(&settings.mailtrap.sender, "mailtrap-sender", "hello@example.com", "Mailtrap Sender")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Get DB connection string from envrc (DSN)
	dsn := os.Getenv("DSN")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("cannot open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test DB connection
	if err := db.Ping(); err != nil {
		logger.Error("cannot connect to database", "error", err)
		os.Exit(1)
	}

	appInstance := &applicationDependencies{
		config:          settings,
		logger:          logger,
		db:              db,
		userModel:       data.UserModel{DB: db},
		tokenModel:      data.TokenModel{DB: db},
		permissionModel: data.PermissionModel{DB: db},
		mailer:          mailer.New(settings.mailtrap.apiKey, settings.mailtrap.sender, settings.mailtrap.url),
	}

	// Use the routes() function from routes.go
	router := appInstance.routes()

	apiServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", settings.port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.Info("shutting down server", "signal", s.String())
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := apiServer.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		appInstance.logger.Info("completing background tasks", "address", apiServer.Addr)
		appInstance.wg.Wait()
		shutdownError <- nil
	}()

	logger.Info("starting server", "address", apiServer.Addr, "environment", settings.environment)

	err = apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	err = <-shutdownError
	if err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped gracefully")
}
