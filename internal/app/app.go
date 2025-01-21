package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/A-PseudoCode-A/grpc_sso/internal/app/grpc"
)

type App struct {
	GRPCSerer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	gRPCApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSerer: gRPCApp,
	}
}
