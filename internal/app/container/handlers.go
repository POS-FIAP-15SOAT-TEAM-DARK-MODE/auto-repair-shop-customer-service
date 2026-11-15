package container

import (
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	customerInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces"

	"github.com/gin-gonic/gin"
)

type HandlersWrapper struct {
	UserHandler             authInterfaces.AuthHTTPController
	CustomerHandler         customerInterfaces.CustomerHTTPController
	VehicleHandler          vehicleInterfaces.VehicleHTTPController
	CustomerInternalHandler customerInterfaces.CustomerInternalHTTPController
	VehicleInternalHandler  vehicleInterfaces.VehicleInternalHTTPController
}

type Middlewares map[string]gin.HandlerFunc
