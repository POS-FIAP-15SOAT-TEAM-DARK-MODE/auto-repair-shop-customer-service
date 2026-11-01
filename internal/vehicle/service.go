package vehicle

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces"
	"golang.org/x/sync/errgroup"
)

type vehicleService struct {
	uow         uow.Executor
	repo        interfaces.VehicleRepository
	idGenerator id.Generator
}

func NewService(uow uow.Executor, repo interfaces.VehicleRepository, idGenerator id.Generator) *vehicleService {
	return &vehicleService{
		uow:         uow,
		repo:        repo,
		idGenerator: idGenerator,
	}
}

func (svc *vehicleService) Create(ctx context.Context, vehicleAdapter adapters.CreateVehicle) (adapters.VehicleResponse, error) {
	vehicle := domain.NewVehicle(
		svc.idGenerator.NewUUID(),
		vehicleAdapter.LicensePlate,
		vehicleAdapter.Brand,
		vehicleAdapter.Model,
		vehicleAdapter.CustomerId,
		vehicleAdapter.Year)

	if err := vehicle.Validate(); err != nil {
		return adapters.VehicleResponse{}, err
	}

	list, err := svc.repo.Search(ctx, adapters.ListVehiclesParams{
		Plate:    vehicle.LicensePlate,
		PageSize: 1,
		Page:     1,
	})
	if err != nil {
		return adapters.VehicleResponse{}, err
	}

	if len(list) > 0 {
		return adapters.VehicleResponse{}, app.ErrDataConflict
	}

	err = svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Save(transactionContext, vehicle)
	})
	if err != nil {
		return adapters.VehicleResponse{}, err
	}

	return adapters.VehicleDomainToResponse(*vehicle), nil
}

func (svc *vehicleService) List(ctx context.Context, params adapters.ListVehiclesParams) (adapters.PaginatedVehicleResponse, error) {
	if params.CustomerID == "" && params.Plate == "" {
		return adapters.PaginatedVehicleResponse{}, domain.ErrInvalidSearchVehicleParams
	}

	if params.Plate != "" {
		params.Plate = domain.NormalizeLicensePlate(params.Plate)
	}

	var totalItems int64
	var items []domain.Vehicle
	eg, groupCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		count, err := svc.repo.Count(groupCtx, params)
		if err != nil {
			return err
		}
		totalItems = count
		return nil
	})

	eg.Go(func() error {
		vehicles, err := svc.repo.Search(groupCtx, params)
		if err != nil {
			return err
		}
		items = vehicles
		return nil
	})

	if err := eg.Wait(); err != nil {
		logger.Of(ctx).Error(fmt.Errorf("error listing vehicles: %w", err))
		return adapters.PaginatedVehicleResponse{}, err
	}

	return adapters.PaginatedVehicleResponse{
		Items:      adapters.VehiclesDomainToResponse(items),
		TotalItems: totalItems,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: int64(math.Max(float64(totalItems/params.PageSize), 1)),
	}, nil
}

func (svc *vehicleService) Edit(ctx context.Context, id string, vehicleAdapter adapters.CreateVehicle) (adapters.VehicleResponse, error) {
	if id == "" {
		return adapters.VehicleResponse{}, domain.ErrInvalidVehicleId
	}

	vehicle := domain.NewVehicle(
		id,
		vehicleAdapter.LicensePlate,
		vehicleAdapter.Brand,
		vehicleAdapter.Model,
		vehicleAdapter.CustomerId,
		vehicleAdapter.Year)

	if err := vehicle.Validate(); err != nil {
		return adapters.VehicleResponse{}, err
	}

	list, err := svc.repo.Search(ctx, adapters.ListVehiclesParams{VehicleID: id, Page: 1, PageSize: 1})
	if err != nil {
		return adapters.VehicleResponse{}, err
	}

	if len(list) == 0 {
		return adapters.VehicleResponse{}, domain.ErrVehicleNotFound
	}

	err = svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Save(transactionContext, vehicle)
	})
	if err != nil {
		return adapters.VehicleResponse{}, err
	}

	return adapters.VehicleDomainToResponse(*vehicle), nil
}

func (svc *vehicleService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidVehicleId
	}

	return svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Delete(transactionContext, id)
	})
}

func (svc *vehicleService) FindById(ctx context.Context, id string) (domain.Vehicle, error) {
	if id == "" {
		return domain.Vehicle{}, domain.ErrInvalidVehicleId
	}

	vehicles, err := svc.repo.Search(ctx, adapters.ListVehiclesParams{VehicleID: id, Page: 1, PageSize: 1})
	if err != nil {
		return domain.Vehicle{}, err
	}

	if len(vehicles) == 0 {
		return domain.Vehicle{}, domain.ErrVehicleNotFound
	}

	return vehicles[0], nil
}
