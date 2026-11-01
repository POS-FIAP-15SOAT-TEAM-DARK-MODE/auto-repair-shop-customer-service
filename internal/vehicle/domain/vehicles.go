package domain

import (
	"errors"
	"regexp"
	"strings"
)

func NewVehicle(id, plate, brand, model, customerId string, year int) *Vehicle {
	return &Vehicle{
		ID:           id,
		LicensePlate: strings.ToUpper(plate),
		Brand:        brand,
		Model:        model,
		Year:         year,
		CustomerId:   customerId,
	}
}

type Vehicle struct {
	ID           string
	LicensePlate string
	Brand        string
	Model        string
	Year         int
	CustomerId   string
}

const firstCarYear = 1886

var platePattern = regexp.MustCompile(`^[A-Za-z]{3}[0-9][A-Za-z0-9][0-9]{2}$`)

func (v *Vehicle) Validate() error {
	var errs []error

	if err := v.validateLicensePlate(); err != nil {
		errs = append(errs, err)
	}

	if err := v.validateYear(); err != nil {
		errs = append(errs, err)
	}

	if err := v.validateBrand(); err != nil {
		errs = append(errs, err)
	}

	if err := v.validateModel(); err != nil {
		errs = append(errs, err)
	}

	if err := v.validateOwner(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (v *Vehicle) validateLicensePlate() error {
	if strings.TrimSpace(v.LicensePlate) == "" {
		return ErrVehicleInvalidPlate
	}

	v.LicensePlate = NormalizeLicensePlate(v.LicensePlate)
	if !platePattern.MatchString(v.LicensePlate) {
		return ErrVehicleInvalidPlate
	}
	return nil
}

func NormalizeLicensePlate(licensePlate string) string {
	licensePlate = strings.ReplaceAll(licensePlate, " ", "")
	return strings.ToUpper(strings.ReplaceAll(licensePlate, "-", ""))
}

func (v *Vehicle) validateYear() error {
	if v.Year < firstCarYear {
		return ErrParamVehicleYear
	}
	return nil
}

func (v *Vehicle) validateBrand() error {
	if strings.TrimSpace(v.Brand) == "" {
		return ErrRequiredVehicleBrand
	}
	return nil
}

func (v *Vehicle) validateModel() error {
	if strings.TrimSpace(v.Model) == "" {
		return ErrRequiredVehicleModel
	}
	return nil
}

func (v *Vehicle) validateOwner() error {
	if strings.TrimSpace(v.CustomerId) == "" {
		return ErrVehicleNoCustomerAssociated
	}
	return nil
}
