package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	httpapp "sso/internal/app/http"
	"sso/internal/services/auth"
	"sso/internal/storage/sqlite"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
	HTTPSrv *httpapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	httPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)
	httpApp := httpapp.New(log, authService, httPort)

	return &App{
		GRPCSrv: grpcApp,
		HTTPSrv: httpApp,
	}
}
