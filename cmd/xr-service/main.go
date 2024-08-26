package main

import (
	"context"
	"os/signal"
	"syscall"

	xrserver "github.com/Memonagi/wallet_project/internal/xr-server"
	xrservice "github.com/Memonagi/wallet_project/internal/xr-service"
	"github.com/sirupsen/logrus"
)

const port = 2607

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	svc := xrservice.New()
	server := xrserver.New(port, svc)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
