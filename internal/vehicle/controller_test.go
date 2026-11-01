package vehicle_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces/mocks"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name: "success",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customerId": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("adapters.CreateVehicle")).
					Return(adapters.VehicleResponse{}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error (missing fields)",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "",
				"model": "",
				"year": 2020,
				"customerId": ""
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("adapters.CreateVehicle")).
					Return(adapters.VehicleResponse{}, domain.ErrRequiredVehicleBrand)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `{"licensePlate":`,
			mockSetup:      func(m *mocks.VehicleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customerId": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("adapters.CreateVehicle")).
					Return(adapters.VehicleResponse{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := vehicle.NewController(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.Create, http.StatusCreated)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const plateParam = "plate"

	tests := []struct {
		name           string
		param          string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name:  "success",
			param: "abc-1d23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{
						Items: []adapters.VehicleResponse{{
							ID:           "1",
							LicensePlate: "ABC1D23",
							BrandModel:   "chevrolet - onix",
							Year:         2020,
							CustomerId:   "123",
						}},
						TotalItems: 1,
						TotalPages: 1,
						PageSize:   10,
						Page:       1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "not found error",
			param: "ABC1D23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{}, domain.ErrVehicleNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "service error",
			param: "ABC1D23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{}, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "plate normalization",
			param: "  abc-1d23  ",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{
						Items: []adapters.VehicleResponse{{
							ID:           "1",
							LicensePlate: "ABC1D23",
							BrandModel:   "chevrolet - onix",
							Year:         2020,
							CustomerId:   "123",
						}},
						TotalItems: 1,
						TotalPages: 1,
						PageSize:   10,
						Page:       1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "success - should return vehicles",
			param: "123",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{
						Items: []adapters.VehicleResponse{{
							ID:           "1",
							LicensePlate: "ABC1D23",
							BrandModel:   "chevrolet - onix",
							Year:         2020,
							CustomerId:   "123",
						}},
						TotalItems: 1,
						TotalPages: 1,
						PageSize:   10,
						Page:       1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error - invalid query params",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListVehiclesParams")).
					Return(adapters.PaginatedVehicleResponse{}, domain.ErrInvalidSearchVehicleParams)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := vehicle.NewController(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// set query parameter ?plate=...
			req, _ := http.NewRequest(http.MethodGet, "/?"+plateParam+"="+tt.param, nil)
			c.Request = req

			routing.GinHandler(h.List)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_Edit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const idParam = "id"

	tests := []struct {
		name           string
		id             string
		body           string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "veh-1",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customerId": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Edit(mock.Anything, "veh-1", mock.MatchedBy(func(v adapters.CreateVehicle) bool {
						return v.LicensePlate == "ABC1D23"
					})).
					Return(adapters.VehicleResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing id",
			id:   "",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customerId": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Edit(mock.Anything, "", mock.MatchedBy(func(v adapters.CreateVehicle) bool {
						return v.LicensePlate == "ABC1D23"
					})).
					Return(adapters.VehicleResponse{}, domain.ErrInvalidVehicleId)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error (missing fields)",
			id:   "veh-1",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "",
				"model": "",
				"year": 2020,
				"customerId": ""
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Edit(mock.Anything, "veh-1", mock.MatchedBy(func(v adapters.CreateVehicle) bool {
						return v.LicensePlate == "ABC1D23"
					})).
					Return(adapters.VehicleResponse{}, domain.ErrRequiredVehicleBrand)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			id:             "veh-1",
			body:           `{"licensePlate":`,
			mockSetup:      func(m *mocks.VehicleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			id:   "veh-1",
			body: `{
				"licensePlate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customerId": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Edit(mock.Anything, "veh-1", mock.AnythingOfType("adapters.CreateVehicle")).
					Return(adapters.VehicleResponse{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := vehicle.NewController(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: idParam, Value: tt.id},
			}

			req, _ := http.NewRequest(http.MethodPut, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.Edit)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const idParam = "id"

	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name: "missing id",
			id:   "",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Delete(mock.Anything, "").
					Return(domain.ErrInvalidVehicleId)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			id:   "veh-1",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Delete(mock.Anything, "veh-1").
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "deleted with success",
			id:   "veh-1",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Delete(mock.Anything, "veh-1").
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := vehicle.NewController(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: idParam, Value: tt.id},
			}

			req, _ := http.NewRequest(http.MethodPut, "/", nil)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinOnlyErrorHandler(h.Delete)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
