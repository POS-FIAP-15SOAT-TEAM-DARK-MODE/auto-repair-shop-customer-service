package main

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/bootstrap"
	pkgApp "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
)

func main() {
	defer logger.Global().Sync()

	if err := auth.RequireSecret(); err != nil {
		logger.Global().Fatal("config.invalid: JWT_SECRET is required")
	}

	httpServer := bootstrap.HTTPServer()
	manager := pkgApp.NewLifecycleManager(5 * time.Second)
	manager.Add(httpServer)

	if err := manager.Run(context.Background()); err != nil {
		logger.Global().Fatal("application failed")
	}
}
