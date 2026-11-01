package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVehicle(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		plate         string
		brand         string
		model         string
		customerID    string
		year          int
		expectedPlate string
	}{
		{
			name:          "uppercases license plate",
			id:            "vehicle-id",
			plate:         "abc1d23",
			brand:         "chevrolet",
			model:         "onix",
			customerID:    "customer-1",
			year:          2020,
			expectedPlate: "ABC1D23",
		},
		{
			name:          "preserves already uppercase plate",
			id:            "vehicle-id-2",
			plate:         "ABC1D23",
			brand:         "fiat",
			model:         "uno",
			customerID:    "customer-2",
			year:          2010,
			expectedPlate: "ABC1D23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewVehicle(tt.id, tt.plate, tt.brand, tt.model, tt.customerID, tt.year)

			assert.Equal(t, tt.id, v.ID)
			assert.Equal(t, tt.expectedPlate, v.LicensePlate)
			assert.Equal(t, tt.brand, v.Brand)
			assert.Equal(t, tt.model, v.Model)
			assert.Equal(t, tt.year, v.Year)
			assert.Equal(t, tt.customerID, v.CustomerId)
		})
	}
}

func TestNormalizeLicensePlate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "already normalized", input: "ABC1D23", expected: "ABC1D23"},
		{name: "lowercase plate", input: "abc1d23", expected: "ABC1D23"},
		{name: "with hyphen", input: "ABC-1D23", expected: "ABC1D23"},
		{name: "with spaces", input: "  ABC1D23  ", expected: "ABC1D23"},
		{name: "lowercase with spaces and hyphen", input: "  abc-1d23  ", expected: "ABC1D23"},
		{name: "multiple hyphens", input: "A-B-C-1-D-2-3", expected: "ABC1D23"},
		{name: "empty string", input: "", expected: ""},
		{name: "only spaces", input: "     ", expected: ""},
		{name: "only hyphens", input: "-----", expected: ""},
		{name: "mixed format with spaces and hyphens", input: " a-b c-1d 2-3 ", expected: "ABC1D23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeLicensePlate(tt.input))
		})
	}
}

func TestVehicle_Validate(t *testing.T) {
	validVehicle := func() *Vehicle {
		return &Vehicle{
			LicensePlate: "ABC1D23",
			Brand:        "chevrolet",
			Model:        "onix",
			Year:         2020,
			CustomerId:   "customer-1",
		}
	}

	tests := []struct {
		name        string
		vehicle     *Vehicle
		expectedErr error
	}{
		{
			name:        "valid vehicle",
			vehicle:     validVehicle(),
			expectedErr: nil,
		},
		{
			name: "invalid plate",
			vehicle: func() *Vehicle {
				v := validVehicle()
				v.LicensePlate = "INVALID"
				return v
			}(),
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name: "invalid year",
			vehicle: func() *Vehicle {
				v := validVehicle()
				v.Year = 1800
				return v
			}(),
			expectedErr: ErrParamVehicleYear,
		},
		{
			name: "empty brand",
			vehicle: func() *Vehicle {
				v := validVehicle()
				v.Brand = "   "
				return v
			}(),
			expectedErr: ErrRequiredVehicleBrand,
		},
		{
			name: "empty model",
			vehicle: func() *Vehicle {
				v := validVehicle()
				v.Model = ""
				return v
			}(),
			expectedErr: ErrRequiredVehicleModel,
		},
		{
			name: "missing customer",
			vehicle: func() *Vehicle {
				v := validVehicle()
				v.CustomerId = ""
				return v
			}(),
			expectedErr: ErrVehicleNoCustomerAssociated,
		},
		{
			name: "multiple validation errors",
			vehicle: &Vehicle{
				LicensePlate: "XXX",
				Brand:        "",
				Model:        "",
				Year:         1000,
				CustomerId:   "",
			},
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name: "plate normalized with spaces and hyphen",
			vehicle: &Vehicle{
				LicensePlate: " abc-1d23 ",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			expectedErr: nil,
		},
		{
			name: "legacy plate format normalized to valid mercosul pattern",
			vehicle: &Vehicle{
				LicensePlate: "ABC-1234",
				Brand:        "fiat",
				Model:        "palio",
				Year:         2015,
				CustomerId:   "customer-1",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vehicle.Validate()

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.True(t, errors.Is(err, tt.expectedErr),
				"expected error to wrap %v, got %v", tt.expectedErr, err)
		})
	}
}
