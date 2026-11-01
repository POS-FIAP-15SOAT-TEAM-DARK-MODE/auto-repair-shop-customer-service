package repository

import (
	"context"
	"slices"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

type inMemory struct {
	data map[string]domain.Vehicle
}

func NewInMemory() *inMemory {
	return &inMemory{
		make(map[string]domain.Vehicle),
	}
}

func (i *inMemory) Save(_ context.Context, vehicle *domain.Vehicle) error {
	i.data[vehicle.ID] = *vehicle
	return nil
}

func (i *inMemory) Delete(_ context.Context, id string) error {
	delete(i.data, id)
	return nil
}

func (i *inMemory) Search(_ context.Context, params adapters.ListVehiclesParams) ([]domain.Vehicle, error) {
	var filtered []domain.Vehicle
	for _, v := range i.data {
		if params.CustomerID != "" && v.CustomerId != params.CustomerID {
			continue
		}

		if params.Plate != "" && v.LicensePlate != params.Plate {
			continue
		}

		if params.VehicleID != "" && v.ID != params.VehicleID {
			continue
		}
		filtered = append(filtered, v)
	}
	slices.SortFunc(filtered, func(a, b domain.Vehicle) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	start := int((params.Page - 1) * params.PageSize)
	if start >= len(filtered) {
		return []domain.Vehicle{}, nil
	}
	end := start + int(params.PageSize)
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], nil
}

func (i *inMemory) Count(_ context.Context, params adapters.ListVehiclesParams) (int64, error) {
	var total int64
	for _, v := range i.data {
		if params.CustomerID != "" && v.CustomerId != params.CustomerID {
			continue
		}

		if params.Plate != "" && v.LicensePlate != params.Plate {
			continue
		}

		total++
	}
	return total, nil
}
