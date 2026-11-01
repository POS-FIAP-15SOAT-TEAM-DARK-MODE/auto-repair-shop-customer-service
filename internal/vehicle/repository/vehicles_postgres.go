package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	postgres "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"go.uber.org/zap"
)

const (
	upsertNewVehicle = `
	INSERT INTO "vehicle" (id, license_plate, brand, model, year, customer_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE
	SET license_plate = EXCLUDED.license_plate,
	    brand = EXCLUDED.brand,
	    model = EXCLUDED.model,
	    year = EXCLUDED.year,
	    customer_id = EXCLUDED.customer_id,
	    updated_at = NOW();`

	selectVehicle = `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle`

	deleteVehicle = `DELETE FROM vehicle WHERE id = $1`

	countVehicles = `SELECT COUNT(id) FROM vehicle`
)

type postgresRepository struct{}

func NewPostgres() *postgresRepository {
	return &postgresRepository{}
}

func (r *postgresRepository) Save(ctx context.Context, vehicle *domain.Vehicle) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	affected, err := tx.ExecContext(
		ctx,
		upsertNewVehicle,
		vehicle.ID,
		vehicle.LicensePlate,
		vehicle.Brand,
		vehicle.Model,
		vehicle.Year,
		vehicle.CustomerId,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	if amount, e := affected.RowsAffected(); amount == 0 {
		if e != nil {
			return pgPkg.Error(ctx, e)
		}

		return domain.ErrVehicleNotFound
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, deleteVehicle, id); err != nil {
		logger.Of(ctx).Warn("vehicle repository", zap.String("operation", "delete"), zap.Error(err))
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *postgresRepository) Search(ctx context.Context, params adapters.ListVehiclesParams) ([]domain.Vehicle, error) {
	runner, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.PageSize
	offset := (params.Page - 1) * params.PageSize
	qb := db.QueryBuilder(selectVehicle).AddPagination(limit, offset)
	if params.CustomerID != "" {
		qb.Add("customer_id =", params.CustomerID)
	}
	if params.Plate != "" {
		qb.Add("license_plate =", params.Plate)
	}
	if params.VehicleID != "" {
		qb.Add("id =", params.VehicleID)
	}
	query, args := qb.Build()

	rows, err := runner.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	vehicles := make([]domain.Vehicle, 0, limit)
	for rows.Next() {
		var v domain.Vehicle
		if err = rows.Scan(
			&v.ID,
			&v.LicensePlate,
			&v.Brand,
			&v.Model,
			&v.Year,
			&v.CustomerId,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		vehicles = append(vehicles, v)
	}
	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return vehicles, nil
}

func (r *postgresRepository) Count(ctx context.Context, params adapters.ListVehiclesParams) (int64, error) {
	runner, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	qb := db.QueryBuilder(countVehicles)
	if params.CustomerID != "" {
		qb.Add("customer_id =", params.CustomerID)
	}
	if params.Plate != "" {
		qb.Add("license_plate =", params.Plate)
	}
	if params.VehicleID != "" {
		qb.Add("id =", params.VehicleID)
	}
	query, args := qb.Build()

	var total int64
	if err = runner.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}
