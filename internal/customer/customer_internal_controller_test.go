package customer

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func internalCtx(id string) context.Context {
	return context.WithValue(context.Background(), "id", id) //nolint:staticcheck // gin param keys are plain strings by design
}

func TestGetInternalByID_IndividualHasNullCompanyName(t *testing.T) {
	svc := mocks.NewCustomerService(t)
	svc.EXPECT().GetByID(mock.Anything, "c1").Return(adapters.CustomerResponse{
		ID:           "c1",
		Type:         "INDIVIDUAL",
		Document:     "52998224725",
		Phone:        "11999990001",
		CompanyName:  "",
		UserResponse: authAdapters.UserResponse{ID: "u1", Name: "João Silva", Email: "joao.silva@email.com"},
	}, nil)

	got, err := NewController(svc).GetInternalByID(internalCtx("c1"), &http.Request{})
	require.NoError(t, err)
	assert.Equal(t, adapters.InternalCustomerResponse{
		ID: "c1", UserID: "u1", Type: "INDIVIDUAL", Name: "João Silva", Email: "joao.silva@email.com",
		Document: "52998224725", CompanyName: nil, Phone: "11999990001",
	}, got)
}

func TestGetInternalByID_CompanyKeepsUserNameAndSetsCompanyName(t *testing.T) {
	svc := mocks.NewCustomerService(t)
	svc.EXPECT().GetByID(mock.Anything, "c2").Return(adapters.CustomerResponse{
		ID:           "c2",
		Type:         domain.CompanyCustomerType.String(),
		Document:     "11.222.333/0001-81",
		CompanyName:  "Transportes Silva Ltda",
		Phone:        "11988880002",
		UserResponse: authAdapters.UserResponse{ID: "u2", Name: "Auto Transportes Silva", Email: "contato@t.com.br"},
	}, nil)

	got, err := NewController(svc).GetInternalByID(internalCtx("c2"), &http.Request{})
	require.NoError(t, err)
	assert.Equal(t, "Auto Transportes Silva", got.Name)
	require.NotNil(t, got.CompanyName)
	assert.Equal(t, "Transportes Silva Ltda", *got.CompanyName)
	assert.Equal(t, "contato@t.com.br", got.Email)
	assert.Equal(t, "11988880002", got.Phone)
	assert.Equal(t, "11222333000181", got.Document)
	assert.Equal(t, "COMPANY", got.Type)
}

func TestGetInternalByID_PropagatesNotFound(t *testing.T) {
	svc := mocks.NewCustomerService(t)
	svc.EXPECT().GetByID(mock.Anything, "missing").Return(adapters.CustomerResponse{}, domain.ErrCustomerNotFound)

	_, err := NewController(svc).GetInternalByID(internalCtx("missing"), &http.Request{})
	assert.True(t, errors.Is(err, domain.ErrCustomerNotFound))
}

func TestUnformatDocument(t *testing.T) {
	assert.Equal(t, "52998224725", unformatDocument("529.982.247-25"))
	assert.Equal(t, "45997418000153", unformatDocument("45.997.418/0001-53"))
	assert.Equal(t, "", unformatDocument(""))
}
