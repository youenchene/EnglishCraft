package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/youenchene/EnglishCraft/internal/curriculum/adapters/content"
	"github.com/youenchene/EnglishCraft/internal/identity/adapters/password"
	identityadapter "github.com/youenchene/EnglishCraft/internal/identity/adapters/postgres"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	httpapi "github.com/youenchene/EnglishCraft/internal/learning/adapters/httpapi"
	learningpostgres "github.com/youenchene/EnglishCraft/internal/learning/adapters/postgres"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	"github.com/youenchene/EnglishCraft/internal/platform/config"
	platformpostgres "github.com/youenchene/EnglishCraft/internal/platform/postgres"
	"github.com/youenchene/EnglishCraft/internal/platform/system"
)

func Run(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("parse database configuration: %w", err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	if os.Getenv("AUTO_MIGRATE") != "false" || os.Getenv("APP_COMMAND") == "migrate" {
		if err := platformpostgres.Migrate(ctx, pool); err != nil {
			return err
		}
	} else {
		var migrated bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=4)`).Scan(&migrated); err != nil || !migrated {
			return errors.New("database migrations required; run APP_COMMAND=migrate with the new image")
		}
	}

	store := learningpostgres.NewStore(pool)
	catalog, err := content.Load()
	if err != nil {
		return fmt.Errorf("load content: %w", err)
	}
	service := core.New(catalog, store, system.Clock{})
	sessions := identity.NewSessions(identityadapter.New(pool), password.Bcrypt{}, system.Clock{})
	if os.Getenv("APP_COMMAND") == "migrate" {
		return nil
	}
	if os.Getenv("APP_COMMAND") == "provision" {
		return sessions.Provision(ctx, os.Getenv("INITIAL_PARENT_USERNAME"), os.Getenv("INITIAL_PARENT_PASSWORD"),
			os.Getenv("INITIAL_CHILD_USERNAME"), os.Getenv("INITIAL_CHILD_PASSWORD"))
	}

	apiMux := http.NewServeMux()
	tutor, imports, upstream, err := tutoring(pool, service)
	if err != nil {
		return fmt.Errorf("configure tutoring: %w", err)
	}
	handler := httpapi.NewCore(service, sessions, settings.CookieSecure)
	handler.Register(apiMux)
	handler.RegisterTutoring(apiMux, tutor, imports)
	handler.RegisterDeletion(apiMux, identity.NewDeletion(sessions, identityadapter.New(pool)), tutor)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				deadline, cancel := context.WithTimeout(ctx, 5*time.Second)
				if err := tutor.ReapVoice(deadline); err != nil {
					logger.Warn("voice cleanup failed")
				}
				cancel()
			}
		}
	}()
	rootMux := http.NewServeMux()
	rootMux.HandleFunc("GET /healthz", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})
	rootMux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		deadline, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if pool.Ping(deadline) != nil {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(204)
	})
	rootMux.Handle("/api/", httpapi.SameOrigin(httpapi.SensitiveRateLimit(apiMux)))
	if upstream != nil {
		rootMux.Handle("GET /api/v1/tutor/voice/elevenlabs/upstream", upstream)
	}
	rootMux.Handle("/", httpapi.WebApp())

	server := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              settings.Address,
		Handler:           httpapi.Recover(logger, httpapi.AccessLog(logger, httpapi.SecurityHeaders(rootMux))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("EnglishCraft started", "address", settings.Address)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), settings.ShutdownGrace)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	}
}
