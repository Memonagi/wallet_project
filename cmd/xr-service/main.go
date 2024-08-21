package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/config"
	xrserver "github.com/Memonagi/wallet_project/internal/xr-server"
	xrservice "github.com/Memonagi/wallet_project/internal/xr-service"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	cfg := config.New()

	svc := xrservice.New()
	server := xrserver.New(xrserver.Config{Port: cfg.GetXRPort()}, svc)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
