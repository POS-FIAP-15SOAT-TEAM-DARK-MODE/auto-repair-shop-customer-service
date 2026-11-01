package vehicle

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	uowMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	domainMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository)
		vehicle    adapters.CreateVehicle
		wantResp   adapters.VehicleResponse
		wantErr    error
	}{
		{
			name: "Success",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					Plate:    "ABC1D23",
					PageSize: 1,
					Page:     1,
				}).Return(nil, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					return steps[0](context.Background())
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{
				ID:           "mock-id",
				LicensePlate: "ABC1D23",
				BrandModel:   "chevrolet - onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantErr: nil,
		},
		{
			name: "ValidationError",
			vehicle: adapters.CreateVehicle{
				LicensePlate: "INVALID", // inválida
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  errors.Join(domain.ErrVehicleInvalidPlate),
		},
		{
			name: "UowError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					Plate:    "ABC1D23",
					PageSize: 1,
					Page:     1,
				}).Return(nil, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).Return(errors.New("uow error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  errors.New("uow error"),
		},
		{
			name: "RepositoryError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					Plate:    "ABC1D23",
					PageSize: 1,
					Page:     1,
				}).Return(nil, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					return steps[0](context.Background())
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(errors.New("db error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  errors.New("db error"),
		},
		{
			name: "SearchError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					Plate:    "ABC1D23",
					PageSize: 1,
					Page:     1,
				}).Return(nil, errors.New("search error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  errors.New("search error"),
		},
		{
			name: "DataConflictError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				// Simulate that Search returns a non-empty result (license plate already exists)
				existing := []domain.Vehicle{
					{
						ID:           "existing-id",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "456",
					},
				}
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					Plate:    "ABC1D23",
					PageSize: 1,
					Page:     1,
				}).Return(existing, nil)
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  app.ErrDataConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exec := uowMocks.NewExecutor(t)
			repo := domainMocks.NewVehicleRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(exec, repo)
			}
			service := NewService(exec, repo, id.NewNoopIDGenerator("mock-id"))

			resp, err := service.Create(ctx, tt.vehicle)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}

func TestService_List(t *testing.T) {
	expectedItems := []domain.Vehicle{
		{ID: "1", LicensePlate: "AAA1234", Brand: "chevrolet", Model: "onix", Year: 2020, CustomerId: "123"},
	}
	tests := []struct {
		name       string
		params     adapters.ListVehiclesParams
		setupMocks func(repo *domainMocks.VehicleRepository, params adapters.ListVehiclesParams)
		wantErr    error
		wantResp   adapters.PaginatedVehicleResponse
	}{
		{
			name: "Success with CustomerID",
			params: adapters.ListVehiclesParams{
				CustomerID: "123",
				Page:       1,
				PageSize:   10,
			},
			setupMocks: func(repo *domainMocks.VehicleRepository, searchParams adapters.ListVehiclesParams) {
				repo.EXPECT().Count(mock.Anything, searchParams).Return(int64(1), nil)
				repo.EXPECT().Search(mock.Anything, searchParams).Return(expectedItems, nil)
			},
			wantResp: adapters.PaginatedVehicleResponse{
				TotalItems: 1,
				Items: []adapters.VehicleResponse{{
					ID:           "1",
					LicensePlate: "AAA1234",
					BrandModel:   "chevrolet - onix",
					Year:         2020,
					CustomerId:   "123",
				}},
				PageSize:   10,
				Page:       1,
				TotalPages: 1,
			},
		},
		{
			name: "Success with Plate",
			params: adapters.ListVehiclesParams{
				Plate:    "AAA1234",
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(repo *domainMocks.VehicleRepository, searchParams adapters.ListVehiclesParams) {
				repo.EXPECT().Count(mock.Anything, searchParams).Return(int64(1), nil)
				repo.EXPECT().Search(mock.Anything, searchParams).Return(expectedItems, nil)
			},
			wantResp: adapters.PaginatedVehicleResponse{
				TotalItems: 1,
				Items: []adapters.VehicleResponse{{
					ID:           "1",
					LicensePlate: "AAA1234",
					BrandModel:   "chevrolet - onix",
					Year:         2020,
					CustomerId:   "123",
				}},
				PageSize:   10,
				Page:       1,
				TotalPages: 1,
			},
		},
		{
			name: "Success - EmptyResult",
			params: adapters.ListVehiclesParams{
				CustomerID: "123",
				Page:       1,
				PageSize:   10,
			},
			setupMocks: func(repo *domainMocks.VehicleRepository, searchParams adapters.ListVehiclesParams) {
				repo.EXPECT().Count(mock.Anything, searchParams).Return(int64(0), nil)
				repo.EXPECT().Search(mock.Anything, searchParams).Return([]domain.Vehicle{}, nil)
			},
			wantResp: adapters.PaginatedVehicleResponse{
				TotalItems: 0,
				Items:      []adapters.VehicleResponse{},
				PageSize:   10,
				Page:       1,
				TotalPages: 1,
			},
		},
		{
			name: "ValidateSearchParams",
			params: adapters.ListVehiclesParams{
				Page:     2,
				PageSize: 10,
			},
			wantErr: domain.ErrInvalidSearchVehicleParams,
		},
		{
			name: "CountError",
			params: adapters.ListVehiclesParams{
				CustomerID: "123",
				Page:       1,
				PageSize:   10,
			},
			setupMocks: func(repo *domainMocks.VehicleRepository, searchParams adapters.ListVehiclesParams) {
				repo.EXPECT().Count(mock.Anything, searchParams).Return(int64(0), errors.New("count error"))
				repo.EXPECT().Search(mock.Anything, searchParams).Return([]domain.Vehicle{}, nil)
			},
			wantErr: errors.New("count error"),
		},
		{
			name: "SearchError",
			params: adapters.ListVehiclesParams{
				CustomerID: "123",
				Page:       1,
				PageSize:   10,
			},
			setupMocks: func(repo *domainMocks.VehicleRepository, searchParams adapters.ListVehiclesParams) {
				repo.EXPECT().Count(mock.Anything, searchParams).Return(int64(1), nil)
				repo.EXPECT().Search(mock.Anything, searchParams).Return(nil, errors.New("search error"))
			},
			wantErr: errors.New("search error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := domainMocks.NewVehicleRepository(t)
			if tt.setupMocks != nil {
				tt.setupMocks(repo, tt.params)
			}
			service := NewService(nil, repo, nil)
			resp, err := service.List(ctx, tt.params)
			assert.Equal(t, tt.wantResp, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_Edit(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository)
		id         string
		vehicle    adapters.CreateVehicle
		wantResp   adapters.VehicleResponse
		wantErr    error
	}{
		{
			name: "Success",
			id:   "mock-id",
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{
				ID:           "mock-id",
				LicensePlate: "ABC1D23",
				BrandModel:   "chevrolet - onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantErr: nil,
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					VehicleID: "mock-id",
					PageSize:  1,
					Page:      1,
				}).Return([]domain.Vehicle{
					{
						ID:           "mock-id",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "123",
					},
				}, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					return steps[0](context.Background())
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)
			},
		},
		{
			name:    "EmptyIDError",
			wantErr: domain.ErrInvalidVehicleId,
		},
		{
			name: "NotFoundError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					VehicleID: "nonexisting-id",
					PageSize:  1,
					Page:      1,
				}).Return(nil, nil)
			},
			id: "nonexisting-id",
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantResp: adapters.VehicleResponse{},
			wantErr:  domain.ErrVehicleNotFound,
		},
		{
			name: "SearchError",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					VehicleID: "ABC1D23",
					PageSize:  1,
					Page:      1,
				}).Return(nil, errors.New("search error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			id:       "ABC1D23",
			wantResp: adapters.VehicleResponse{},
			wantErr:  errors.New("search error"),
		},
		{
			name: "ValidationError",
			vehicle: adapters.CreateVehicle{
				LicensePlate: "INVALID",
				Year:         2020,
			},
			id:      "ABC1D23",
			wantErr: errors.Join(domain.ErrVehicleInvalidPlate, domain.ErrRequiredVehicleBrand, domain.ErrRequiredVehicleModel, domain.ErrVehicleNoCustomerAssociated),
		},
		{
			name: "UowError",
			id:   "mock-id",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					VehicleID: "mock-id",
					PageSize:  1,
					Page:      1,
				}).Return([]domain.Vehicle{
					{
						ID:           "mock-id",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "123",
					},
				}, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).Return(errors.New("uow error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantErr: errors.New("uow error"),
		},
		{
			name: "RepositoryError",
			id:   "mock-id",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, adapters.ListVehiclesParams{
					VehicleID: "mock-id",
					PageSize:  1,
					Page:      1,
				}).Return([]domain.Vehicle{
					{
						ID:           "mock-id",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "123",
					},
				}, nil)
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					return steps[0](context.Background())
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(errors.New("db error"))
			},
			vehicle: adapters.CreateVehicle{
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			wantErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exec := uowMocks.NewExecutor(t)
			repo := domainMocks.NewVehicleRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(exec, repo)
			}
			service := NewService(exec, repo, nil)

			resp, err := service.Edit(ctx, tt.id, tt.vehicle)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupMocks func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository, id string)
		wantErr    error
	}{
		{
			name: "Success",
			id:   "veh-1",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository, id string) {
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					if err := steps[0](context.Background()); err != nil {
						t.Fatalf("step returned error: %v", err)
					}
					return nil
				})
				repo.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
		},
		{
			name:    "InvalidIdError",
			wantErr: domain.ErrInvalidVehicleId,
		},
		{
			name: "UowError",
			id:   "veh-1",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository, id string) {
				exec.EXPECT().Execute(mock.Anything, mock.Anything).Return(errors.New("uow error"))
			},
			wantErr: errors.New("uow error"),
		},
		{
			name: "RepositoryError",
			id:   "veh-1",
			setupMocks: func(exec *uowMocks.Executor, repo *domainMocks.VehicleRepository, id string) {
				exec.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
					if len(steps) != 1 {
						t.Fatalf("expected 1 step, got %d", len(steps))
					}
					return steps[0](context.Background())
				})
				repo.EXPECT().Delete(mock.Anything, id).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exec := uowMocks.NewExecutor(t)
			repo := domainMocks.NewVehicleRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(exec, repo, tt.id)
			}
			service := NewService(exec, repo, nil)

			err := service.Delete(ctx, tt.id)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestVehicleService_FindById(t *testing.T) {
	t.Parallel()

	someVehicle := domain.Vehicle{ID: "abc-123", Brand: "Toyota", Model: "Corolla"}
	repoErr := errors.New("db connection failed")
	searchParams := adapters.ListVehiclesParams{VehicleID: "abc-123", Page: 1, PageSize: 1}

	tests := []struct {
		name        string
		id          string
		mockSetup   func(repo *domainMocks.VehicleRepository)
		wantVehicle domain.Vehicle
		wantErr     error
	}{
		{
			name: "empty id returns ErrInvalidVehicleId without calling repo",
			id:   "",
			mockSetup: func(repo *domainMocks.VehicleRepository) {
				repo.EXPECT().Search(mock.Anything, mock.Anything).Maybe()
			},
			wantVehicle: domain.Vehicle{},
			wantErr:     domain.ErrInvalidVehicleId,
		},
		{
			name: "repo returns a vehicle, found successfully",
			id:   "abc-123",
			mockSetup: func(repo *domainMocks.VehicleRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Vehicle{someVehicle}, nil).
					Once()
			},
			wantVehicle: someVehicle,
			wantErr:     nil,
		},
		{
			name: "repo returns empty slice, vehicle not found",
			id:   "abc-123",
			mockSetup: func(repo *domainMocks.VehicleRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Vehicle{}, nil).
					Once()
			},
			wantVehicle: domain.Vehicle{},
			wantErr:     domain.ErrVehicleNotFound,
		},
		{
			name: "repo returns an error, propagated as-is",
			id:   "abc-123",
			mockSetup: func(repo *domainMocks.VehicleRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Vehicle{}, repoErr).
					Once()
			},
			wantVehicle: domain.Vehicle{},
			wantErr:     repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exec := uowMocks.NewExecutor(t)
			repo := domainMocks.NewVehicleRepository(t)
			tc.mockSetup(repo)

			svc := NewService(exec, repo, nil)
			got, err := svc.FindById(context.Background(), tc.id)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantVehicle, got)
		})
	}
}
