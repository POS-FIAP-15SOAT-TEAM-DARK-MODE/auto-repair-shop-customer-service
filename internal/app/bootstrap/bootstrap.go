package bootstrap

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
)

func HTTPServer() app.Server {
	middlewares := middlewaresContainer()
	db := app.DBConnect()
	app.RunSeed(context.Background(), db)
	routes := httpContainer(db)
	router := routing.SetupRouter(routes, middlewares)

	port := env.GetString("PORT", "8081")
	return app.NewHTTPServer("http-api", ":"+port, router)
}

func middlewaresContainer() *container.Middlewares {
	return &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"Metrics":      middleware.Metrics(),
		"ErrorHandler": middleware.ErrorHandler(),
	}
}

func httpContainer(db *sql.DB) *container.HandlersWrapper {
	return &container.HandlersWrapper{
		UserHandler:             auth.NewHTTPController(db),
		CustomerHandler:         customer.NewHTTPController(db),
		VehicleHandler:          vehicle.NewHTTPController(db),
		CustomerInternalHandler: customer.NewInternalHTTPController(db),
		VehicleInternalHandler:  vehicle.NewInternalHTTPController(db),
	}
}
