package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"final-project/api/proto/budgetpb"
	"final-project/internal/auth"
	budgetrepo "final-project/internal/repository/budget"
	budgetperiodrepo "final-project/internal/repository/budget_period"
	periodlimitrepo "final-project/internal/repository/period_limit"
	rolerepo "final-project/internal/repository/role"
	userbudgetrolerepo "final-project/internal/repository/user_budget_role"
	userrepo "final-project/internal/repository/user"
	grpctransport "final-project/internal/transport/grpc"

	"final-project/internal/config"
	"final-project/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

func main() {
	// ---------- Config & Logger ----------

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using system env")
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel, cfg.Env)
	slog.SetDefault(log)

	ctx := context.Background()

	// ---------- PostgreSQL ----------

	migrateDB, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		slog.Error("cannot open migration db", "err", err)
		os.Exit(1)
	}

	goose.SetDialect("postgres")
	if err := goose.Up(migrateDB, "migrations"); err != nil {
		slog.Error("migrations failed", "err", err)
		os.Exit(1)
	}
	migrateDB.Close()

	slog.Info("migrations applied successfully")

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("cannot create connection pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// ---------- Repositories ----------

	userRepo := userrepo.NewPostgres(pool)

	// ---------- Handlers ----------

	authHandler := auth.NewHandler(userRepo, cfg.JWTSecret, cfg.JWTAccessTTL)

	// ---------- GRPC server ----------
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		slog.Error("grpc listen", "err", err)
		os.Exit(1)
	}

	budgetRepo := budgetrepo.NewPostgres(pool)
	periodRepo := budgetperiodrepo.NewPostgres(pool)
	periodLimitRepo := periodlimitrepo.NewPostgres(pool)
	roleRepo := rolerepo.NewPostgres(pool)
	userBudgetRoleRepo := userbudgetrolerepo.NewPostgres(pool)
	budgetServer := grpctransport.NewBudgetServer(budgetRepo, periodRepo, periodLimitRepo, roleRepo, userBudgetRoleRepo)

	grpcServer := grpc.NewServer()
	budgetpb.RegisterBudgetServiceServer(grpcServer, budgetServer)

	go func() {
		slog.Info("starting gRPC server", "addr", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("grpc server error", "err", err)
			os.Exit(1)
		}
	}()

	// ---------- HTTP server ----------

	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			log.Info("http request",
				"method", c.Request().Method,
				"path", c.Path(),
				"status", c.Response().Status,
				"dur_ms", time.Since(start).Microseconds(),
			)
			return err
		}
	})

	// Health probes
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	// Auth routes
	api := e.Group("/api/v1/auth")
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	api.GET("/me", authHandler.Me, auth.Middleware([]byte(cfg.JWTSecret)))

	// Start
	slog.Info("starting http server", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := e.Start(cfg.HTTPAddr); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}

}
