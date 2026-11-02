//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	authPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	customerPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	customerRepoImpl "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	vehiclePkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/repository"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
)

var (
	testServer     *httptest.Server
	memUserRepo    authInterfaces.AuthRepository
	seedCustomerID string
	seedVehicleID  string
	seedUserID     string
)

// internalToken is the INTERNAL_TOKEN used by the test server (set before the router is built).
const internalToken = "integration-internal-token"

// noServiceOrders stands in for the Service Order service, which is not running in these tests:
// no customer or vehicle has orders. The client itself is covered by its unit tests.
type noServiceOrders struct{}

func (noServiceOrders) HasCustomerOrders(context.Context, string) (bool, error) { return false, nil }

func (noServiceOrders) HasVehicleOrders(context.Context, string) (bool, error) { return false, nil }

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("INTERNAL_TOKEN", internalToken)

	ctx := context.Background()
	idGen := id.NewIDGenerator()
	noopUoW := uow.NewInMemoryUoW()

	// Auth
	userRepository := authRepo.NewInMemory()
	authSvc := authPkg.NewService(noopUoW, userRepository, idGen)
	authCtrl := authPkg.NewController(authSvc)

	// Customer
	custRepository := customerRepoImpl.NewInMemory()
	custSvc := customerPkg.NewService(noopUoW, custRepository, userRepository, noServiceOrders{})
	custCtrl := customerPkg.NewController(custSvc)

	// Vehicle
	vehicleRepository := vehicleRepo.NewInMemory()
	vehicleSvc := vehiclePkg.NewService(noopUoW, vehicleRepository, idGen, noServiceOrders{})
	vehicleCtrl := vehiclePkg.NewController(vehicleSvc)

	memUserRepo = userRepository
	seedCustomerID, seedVehicleID, seedUserID = seedMemory(ctx, userRepository, custRepository, vehicleRepository)

	handlers := &container.HandlersWrapper{
		UserHandler:             authCtrl,
		CustomerHandler:         custCtrl,
		VehicleHandler:          vehicleCtrl,
		CustomerInternalHandler: custCtrl,
		VehicleInternalHandler:  vehicleCtrl,
	}

	middlewares := &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"ErrorHandler": middleware.ErrorHandler(),
	}

	router := routing.SetupRouter(handlers, middlewares)
	testServer = httptest.NewServer(router)
	defer testServer.Close()

	os.Exit(m.Run())
}

func hashPW(raw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(raw), 4)
	if err != nil {
		panic(fmt.Sprintf("seedMemory: bcrypt failed for %q: %v", raw, err))
	}
	return string(h)
}

// seedMemory returns the IDs of João's customer, his vehicle and his user.
func seedMemory(
	ctx context.Context,
	userRepository authInterfaces.AuthRepository,
	custRepository customerRepo.CustomerRepository,
	vehicleRepository vehicleInterfaces.VehicleRepository,
) (string, string, string) {
	// Staff users
	attendantID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: attendantID, Name: "attendant",
		CoreUser: authDomain.CoreUser{Email: "attendant@autorepairshop.com", Password: hashPW("attendant123")},
	})
	_ = userRepository.SaveUserRole(ctx, attendantID, authDomain.ATTENDANT)

	mechanicID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: mechanicID, Name: "mechanic",
		CoreUser: authDomain.CoreUser{Email: "mechanic@autorepairshop.com", Password: hashPW("mechanic123")},
	})
	_ = userRepository.SaveUserRole(ctx, mechanicID, authDomain.MECHANIC)

	// Customer João Silva (CPF sanitized = digits only)
	joaoUserID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: joaoUserID, Name: "João Silva",
		CoreUser: authDomain.CoreUser{Email: "joao.silva@email.com", Password: hashPW("52998224725")},
	})
	_ = userRepository.SaveUserRole(ctx, joaoUserID, authDomain.CUSTOMER)

	joaoCustomerID := uuid.NewString()
	_ = custRepository.Save(ctx, &customerDomain.Customer{
		ID:     joaoCustomerID,
		UserID: joaoUserID,
		Type:   customerDomain.IndividualCustomerType,
		CPF:    "52998224725",
		Phone:  "11999990001",
		User:   &authDomain.User{ID: joaoUserID, Name: "João Silva", CoreUser: authDomain.CoreUser{Email: "joao.silva@email.com"}},
	})

	// Vehicle ABC-1234
	joaoVehicleID := uuid.NewString()
	_ = vehicleRepository.Save(ctx, &vehicleDomain.Vehicle{
		ID:           joaoVehicleID,
		LicensePlate: "ABC1234",
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         2020,
		CustomerId:   joaoCustomerID,
	})

	return joaoCustomerID, joaoVehicleID, joaoUserID
}
