package main

//cmd: go run cmd\sso\main.go --config=config\config.yaml

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/A-PseudoCode-A/grpc_sso/internal/app"
	"github.com/A-PseudoCode-A/grpc_sso/internal/config"
	"github.com/A-PseudoCode-A/grpc_sso/internal/lib/logger/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//TODO: init cfg
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting app", slog.String("env", cfg.Env))

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
	go application.GRPCSerer.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop
	
	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCSerer.Stop()

	log.Info("application stop")

	//TODO: load gRPC-service
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
