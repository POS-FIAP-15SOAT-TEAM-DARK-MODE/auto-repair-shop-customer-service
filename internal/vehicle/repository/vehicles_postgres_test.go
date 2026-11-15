package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/repository"
	"github.com/stretchr/testify/assert"
)

func TestVehicleRepository_Save(t *testing.T) {
	upsertPattern := `INSERT INTO "vehicle" \(id, license_plate, brand, model, year, customer_id\)\s*VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, v *domain.Vehicle)
		vehicle   *domain.Vehicle
		useTx     bool
		wantErr   error
	}{
		{
			name: "upsert insert success",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Vehicle) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			vehicle: &domain.Vehicle{
				ID:           "id-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "cust-1",
			},
			useTx: true,
		},
		{
			name: "upsert update success",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Vehicle) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			vehicle: &domain.Vehicle{
				ID:           "id-1",
				LicensePlate: "XYZ9K88",
				Brand:        "ford",
				Model:        "ka",
				Year:         2015,
				CustomerId:   "cust-2",
			},
			useTx: true,
		},
		{
			name: "upsert exec error",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Vehicle) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			vehicle: &domain.Vehicle{
				ID:           "id-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "cust-1",
			},
			useTx:   true,
			wantErr: errors.New("exec error"),
		},
		{
			name: "upsert no rows affected",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Vehicle) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			vehicle: &domain.Vehicle{
				ID:           "id-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "cust-1",
			},
			useTx:   true,
			wantErr: domain.ErrVehicleNotFound,
		},
		{
			name:    "missing transaction",
			vehicle: &domain.Vehicle{},
			useTx:   false,
			wantErr: postgresdb.ErrMissingPostgresTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewPostgres()

			if tt.useTx {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				defer func() { _ = db.Close() }()

				uow := postgresdb.NewTransactionalUoW(db)
				if tt.setupMock != nil {
					tt.setupMock(mock, tt.vehicle)
				}

				err = uow.Execute(context.Background(), func(ctx context.Context) error {
					return repo.Save(ctx, tt.vehicle)
				})
				if tt.wantErr != nil {
					assert.Error(t, err)
					if errors.Is(tt.wantErr, domain.ErrVehicleNotFound) || errors.Is(tt.wantErr, postgresdb.ErrMissingPostgresTransaction) {
						assert.ErrorIs(t, err, tt.wantErr)
					} else {
						assert.Contains(t, err.Error(), tt.wantErr.Error())
					}
					return
				}
				assert.NoError(t, err)
				return
			}

			err := repo.Save(context.Background(), tt.vehicle)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVehicleRepository_Delete(t *testing.T) {
	deletePattern := `DELETE FROM vehicle WHERE id = \$1`

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, id string)
		id        string
		useTx     bool
		wantErr   error
	}{
		{
			name:    "delete missing tx",
			id:      "id-1",
			useTx:   false,
			wantErr: postgresdb.ErrMissingPostgresTransaction,
		},
		{
			name: "delete success",
			setupMock: func(mock sqlmock.Sqlmock, id string) {
				mock.ExpectBegin()
				mock.ExpectExec(deletePattern).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			id:    "id-1",
			useTx: true,
		},
		{
			name: "delete exec error",
			setupMock: func(mock sqlmock.Sqlmock, id string) {
				mock.ExpectBegin()
				mock.ExpectExec(deletePattern).
					WithArgs(id).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			id:      "id-1",
			useTx:   true,
			wantErr: errors.New("exec error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewPostgres()

			if tt.useTx {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				defer func() { _ = db.Close() }()

				uow := postgresdb.NewTransactionalUoW(db)
				if tt.setupMock != nil {
					tt.setupMock(mock, tt.id)
				}

				err = uow.Execute(context.Background(), func(ctx context.Context) error {
					return repo.Delete(ctx, tt.id)
				})
				if tt.wantErr != nil {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr.Error())
					return
				}
				assert.NoError(t, err)
				return
			}

			err := repo.Delete(context.Background(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVehicleRepository_Search(t *testing.T) {
	vehicleColumns := []string{"id", "license_plate", "brand", "model", "year", "customer_id"}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		params    adapters.ListVehiclesParams
		wantLen   int
		wantErr   error
		wantFirst *domain.Vehicle
	}{
		{
			name: "search success no filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				Page:     1,
				PageSize: 10,
			},
			wantLen: 1,
			wantFirst: &domain.Vehicle{
				ID:           "vh-1",
				LicensePlate: "AABCEA123",
				Brand:        "fiat",
				Model:        "uno",
				Year:         2010,
				CustomerId:   "cust-1",
			},
		},
		{
			name: "search with customer filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE customer_id = $1 LIMIT $2 OFFSET $3`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("cust-1", int64(5), int64(5)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				CustomerID: "cust-1",
				Page:       2,
				PageSize:   5,
			},
			wantLen: 1,
			wantFirst: &domain.Vehicle{
				ID:           "vh-1",
				LicensePlate: "AABCEA123",
				Brand:        "fiat",
				Model:        "uno",
				Year:         2010,
				CustomerId:   "cust-1",
			},
		},
		{
			name: "search with plate filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE license_plate = $1 LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("AABCEA123", int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				Plate:    "AABCEA123",
				Page:     1,
				PageSize: 10,
			},
			wantLen: 1,
		},
		{
			name: "search with id filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE id = $1 LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("vh-1", int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				VehicleID: "vh-1",
				Page:      1,
				PageSize:  10,
			},
			wantLen: 1,
		},
		{
			name: "search with all filters",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE customer_id = $1 AND license_plate = $2 AND id = $3 LIMIT $4`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("cust-1", "AABCEA123", "vh-1", int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				CustomerID: "cust-1",
				VehicleID:  "vh-1",
				Plate:      "AABCEA123",
				Page:       1,
				PageSize:   10,
			},
			wantLen: 1,
		},
		{
			name: "search empty result",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE license_plate = $1 LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("NOTFOUND", int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns))
			},
			params: adapters.ListVehiclesParams{
				Plate:    "NOTFOUND",
				Page:     1,
				PageSize: 10,
			},
			wantLen: 0,
		},
		{
			name: "search query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnError(errors.New("db error"))
			},
			params: adapters.ListVehiclesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "search scan error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(vehicleColumns).
						AddRow("vh-1", "AABCEA123", "fiat", "uno", "INVALID", "cust-1"))
			},
			params: adapters.ListVehiclesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("sql"),
		},
		{
			name: "search rows error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle LIMIT $1`
				rows := sqlmock.NewRows(vehicleColumns).
					AddRow("vh-1", "AABCEA123", "fiat", "uno", 2010, "cust-1")
				rows.RowError(0, errors.New("row error"))
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(rows)
			},
			params: adapters.ListVehiclesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("row error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer func() { _ = db.Close() }()

			app.ConnectWithDB(db)
			repo := repository.NewPostgres()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			got, err := repo.Search(context.Background(), tt.params)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			if tt.wantFirst != nil && len(got) > 0 {
				assert.Equal(t, *tt.wantFirst, got[0])
			}
		})
	}
}

func TestVehicleRepository_Count(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		params    adapters.ListVehiclesParams
		wantCount int64
		wantErr   error
	}{
		{
			name: "count success no filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			wantCount: 10,
		},
		{
			name: "count with customer filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle WHERE customer_id =`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("cust-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListVehiclesParams{
				CustomerID: "cust-1",
			},
			wantCount: 10,
		},
		{
			name: "count with plate filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle WHERE license_plate =`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("AABCEA123").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListVehiclesParams{
				Plate: "AABCEA123",
			},
			wantCount: 10,
		},
		{
			name: "count with id filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle WHERE id =`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("vh-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListVehiclesParams{
				VehicleID: "vh-1",
			},
			wantCount: 10,
		},
		{
			name: "count with all filters",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle WHERE customer_id = $1 AND license_plate = $2 AND id = $3`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("cust-1", "AABCEA123", "vh-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListVehiclesParams{
				CustomerID: "cust-1",
				VehicleID:  "vh-1",
				Plate:      "AABCEA123",
			},
			wantCount: 10,
		},
		{
			name: "count query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "count scan error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(id) FROM vehicle`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"invalid"}))
				mock.ExpectRollback()
			},
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer func() { _ = db.Close() }()

			app.ConnectWithDB(db)
			repo := repository.NewPostgres()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			count, err := repo.Count(context.Background(), tt.params)
			assert.Equal(t, tt.wantCount, count)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
