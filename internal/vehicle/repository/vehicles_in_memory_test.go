package repository

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedInMemory(t *testing.T, repo *inMemory, vehicles []*domain.Vehicle) {
	t.Helper()
	for _, v := range vehicles {
		require.NoError(t, repo.Save(context.Background(), v))
	}
}

func TestInMemory_Save(t *testing.T) {
	tests := []struct {
		name      string
		vehicle   *domain.Vehicle
		update    func(*domain.Vehicle)
		wantBrand string
		wantModel string
	}{
		{
			name: "persists vehicle",
			vehicle: &domain.Vehicle{
				ID:           "vehicle-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			wantBrand: "chevrolet",
			wantModel: "onix",
		},
		{
			name: "overwrites existing vehicle",
			vehicle: &domain.Vehicle{
				ID:           "vehicle-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			update: func(v *domain.Vehicle) {
				v.Brand = "fiat"
				v.Model = "uno"
			},
			wantBrand: "fiat",
			wantModel: "uno",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			ctx := context.Background()

			require.NoError(t, repo.Save(ctx, tt.vehicle))
			if tt.update != nil {
				tt.update(tt.vehicle)
				require.NoError(t, repo.Save(ctx, tt.vehicle))
			}

			got, err := repo.Search(ctx, adapters.ListVehiclesParams{
				VehicleID: tt.vehicle.ID,
				Page:      1,
				PageSize:  10,
			})
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tt.wantBrand, got[0].Brand)
			assert.Equal(t, tt.wantModel, got[0].Model)
		})
	}
}

func TestInMemory_Delete(t *testing.T) {
	tests := []struct {
		name      string
		seed      *domain.Vehicle
		deleteID  string
		wantCount int64
	}{
		{
			name: "deletes existing vehicle",
			seed: &domain.Vehicle{
				ID:           "vehicle-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			deleteID:  "vehicle-1",
			wantCount: 0,
		},
		{
			name:      "non-existent id does not error",
			deleteID:  "missing-id",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			ctx := context.Background()

			if tt.seed != nil {
				require.NoError(t, repo.Save(ctx, tt.seed))
			}

			err := repo.Delete(ctx, tt.deleteID)
			require.NoError(t, err)

			got, err := repo.Search(ctx, adapters.ListVehiclesParams{
				VehicleID: tt.deleteID,
				Page:      1,
				PageSize:  10,
			})
			require.NoError(t, err)
			assert.Empty(t, got)

			total, err := repo.Count(ctx, adapters.ListVehiclesParams{})
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, total)
		})
	}
}

func TestInMemory_Search(t *testing.T) {
	filterSeed := []*domain.Vehicle{
		{ID: "id-1", LicensePlate: "ABC1D23", Brand: "chevrolet", Model: "onix", Year: 2020, CustomerId: "customer-1"},
		{ID: "id-2", LicensePlate: "XYZ9A87", Brand: "fiat", Model: "uno", Year: 2015, CustomerId: "customer-1"},
		{ID: "id-3", LicensePlate: "DEF4G56", Brand: "vw", Model: "gol", Year: 2018, CustomerId: "customer-2"},
	}

	paginationSeed := []*domain.Vehicle{
		{ID: "id-1", LicensePlate: "AAA1111", Brand: "a", Model: "a", Year: 2020, CustomerId: "customer-1"},
		{ID: "id-2", LicensePlate: "BBB2222", Brand: "b", Model: "b", Year: 2020, CustomerId: "customer-1"},
		{ID: "id-3", LicensePlate: "CCC3333", Brand: "c", Model: "c", Year: 2020, CustomerId: "customer-1"},
	}

	tests := []struct {
		name   string
		seed   []*domain.Vehicle
		params adapters.ListVehiclesParams
		want   []string
	}{
		{
			name: "filter by customer id",
			seed: filterSeed,
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
				Page:       1,
				PageSize:   10,
			},
			want: []string{"id-1", "id-2"},
		},
		{
			name: "filter by plate",
			seed: filterSeed,
			params: adapters.ListVehiclesParams{
				Plate:    "XYZ9A87",
				Page:     1,
				PageSize: 10,
			},
			want: []string{"id-2"},
		},
		{
			name: "filter by vehicle id",
			seed: filterSeed,
			params: adapters.ListVehiclesParams{
				VehicleID: "id-3",
				Page:      1,
				PageSize:  10,
			},
			want: []string{"id-3"},
		},
		{
			name: "combined filters",
			seed: filterSeed,
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
				Plate:      "ABC1D23",
				Page:       1,
				PageSize:   10,
			},
			want: []string{"id-1"},
		},
		{
			name: "no matches",
			seed: filterSeed,
			params: adapters.ListVehiclesParams{
				Plate:    "NOTFOUND",
				Page:     1,
				PageSize: 10,
			},
			want: []string{},
		},
		{
			name: "first page",
			seed: paginationSeed,
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
				Page:       1,
				PageSize:   2,
			},
			want: []string{"id-1", "id-2"},
		},
		{
			name: "second page",
			seed: paginationSeed,
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
				Page:       2,
				PageSize:   2,
			},
			want: []string{"id-3"},
		},
		{
			name: "page beyond range",
			seed: paginationSeed,
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
				Page:       3,
				PageSize:   2,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			seedInMemory(t, repo, tt.seed)

			got, err := repo.Search(context.Background(), tt.params)
			require.NoError(t, err)

			ids := make([]string, len(got))
			for i, v := range got {
				ids[i] = v.ID
			}
			assert.Equal(t, tt.want, ids)
		})
	}
}

func TestInMemory_Count(t *testing.T) {
	seed := []*domain.Vehicle{
		{ID: "id-1", LicensePlate: "ABC1D23", Brand: "chevrolet", Model: "onix", Year: 2020, CustomerId: "customer-1"},
		{ID: "id-2", LicensePlate: "XYZ9A87", Brand: "fiat", Model: "uno", Year: 2015, CustomerId: "customer-1"},
		{ID: "id-3", LicensePlate: "DEF4G56", Brand: "vw", Model: "gol", Year: 2018, CustomerId: "customer-2"},
	}

	tests := []struct {
		name   string
		params adapters.ListVehiclesParams
		want   int64
	}{
		{
			name:   "counts all vehicles without filters",
			params: adapters.ListVehiclesParams{},
			want:   3,
		},
		{
			name: "counts by customer id",
			params: adapters.ListVehiclesParams{
				CustomerID: "customer-1",
			},
			want: 2,
		},
		{
			name: "counts by plate",
			params: adapters.ListVehiclesParams{
				Plate: "DEF4G56",
			},
			want: 1,
		},
		{
			name: "returns zero when no match",
			params: adapters.ListVehiclesParams{
				Plate: "NOTFOUND",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			seedInMemory(t, repo, seed)

			total, err := repo.Count(context.Background(), tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, total)
		})
	}
}
