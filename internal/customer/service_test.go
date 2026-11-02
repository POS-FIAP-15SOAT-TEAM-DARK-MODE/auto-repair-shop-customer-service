package customer

import (
	"context"
	"errors"
	"testing"

	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupService(t *testing.T) (*customerService, context.Context) {
	t.Helper()
	svc := NewService(uow.NewInMemoryUoW(), customerRepo.NewInMemory(), authRepo.NewInMemory(), &fakeOrders{}).(*customerService)
	return svc, context.Background()
}

func TestService_CreateGetUpdateDelete_Individual(t *testing.T) {
	svc, ctx := setupService(t)

	created, err := svc.Create(ctx, adapters.CreateCustomerRequest{
		Name: "João Silva", Email: "joao@email.com", Type: "INDIVIDUAL",
		Document: "529.982.247-25", Phone: "11999990001",
	})
	require.NoError(t, err)
	assert.Equal(t, "52998224725", created.Document)
	assert.Equal(t, "INDIVIDUAL", created.Type)
	require.NotEmpty(t, created.ID)

	got, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "João Silva", got.Name)

	byDoc, err := svc.GetByDocument(ctx, "529.982.247-25")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byDoc.ID)

	updated, err := svc.Update(ctx, created.ID, adapters.UpdateCustomerRequest{Name: "João S.", Email: "j@s.com", Phone: "11000000000"})
	require.NoError(t, err)
	assert.Equal(t, "João S.", updated.Name)
	assert.Equal(t, "11000000000", updated.Phone)

	require.NoError(t, svc.Delete(ctx, created.ID))
	_, err = svc.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)

	// Deleting a missing customer is a no-op (idempotent).
	assert.NoError(t, svc.Delete(ctx, created.ID))
}

func TestService_CreateCompany(t *testing.T) {
	svc, ctx := setupService(t)
	resp, err := svc.Create(ctx, adapters.CreateCustomerRequest{
		Name: "Auto Transportes", Email: "c@t.com.br", Type: "COMPANY",
		Document: "45.997.418/0001-53", CompanyName: "Transportes Ltda", Phone: "11988880002",
	})
	require.NoError(t, err)
	assert.Equal(t, "45997418000153", resp.Document)
	assert.Equal(t, "Transportes Ltda", resp.CompanyName)
}

func TestService_Create_DuplicateDocumentConflicts(t *testing.T) {
	svc, ctx := setupService(t)
	req := adapters.CreateCustomerRequest{Name: "João Silva", Email: "a@a.com", Type: "INDIVIDUAL", Document: "529.982.247-25", Phone: "1"}
	_, err := svc.Create(ctx, req)
	require.NoError(t, err)

	req.Email = "b@b.com"
	_, err = svc.Create(ctx, req)
	assert.Error(t, err)
}

func TestService_Create_InvalidDocument(t *testing.T) {
	svc, ctx := setupService(t)
	_, err := svc.Create(ctx, adapters.CreateCustomerRequest{Name: "João Silva", Email: "a@a.com", Type: "INDIVIDUAL", Document: "123", Phone: "1"})
	assert.Error(t, err)
}

func TestService_GetByID_And_Document_Validation(t *testing.T) {
	svc, ctx := setupService(t)

	_, err := svc.GetByID(ctx, "")
	assert.ErrorIs(t, err, domain.ErrInvalidCustomerId)

	_, err = svc.GetByID(ctx, "missing")
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)

	_, err = svc.GetByDocument(ctx, "abc")
	assert.ErrorIs(t, err, domain.ErrInvalidDocumentFormat)

	_, err = svc.GetByDocument(ctx, "529.982.247-25")
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

func TestService_Update_NotFound(t *testing.T) {
	svc, ctx := setupService(t)
	_, err := svc.Update(ctx, "missing", adapters.UpdateCustomerRequest{Name: "X"})
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

// failingUoW is a hand-written executor whose steps never run, used to reach the error path after a lookup.
type failingUoW struct{}

func (failingUoW) Execute(context.Context, ...uow.Step) error {
	return errors.New("unit of work unavailable")
}

func TestService_Delete_UnitOfWorkFailureKeepsCustomer(t *testing.T) {
	svc, ctx := setupService(t)
	created, err := svc.Create(ctx, adapters.CreateCustomerRequest{
		Name: "Maria Souza", Email: "maria@email.com", Type: "INDIVIDUAL",
		Document: "529.982.247-25", Phone: "11999990003",
	})
	require.NoError(t, err)

	failing := NewService(failingUoW{}, svc.customerRepo, svc.authRepo, &fakeOrders{})
	err = failing.Delete(ctx, created.ID)
	assert.EqualError(t, err, "unit of work unavailable")

	got, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

// listFailingRepo keeps the real Save/Delete of the in-memory repository but fails every List call.
type listFailingRepo struct {
	interfaces.CustomerRepository
}

func (listFailingRepo) List(context.Context, adapters.ListCustomerParams) ([]domain.Customer, error) {
	return nil, errors.New("database unavailable")
}

func TestService_LookupFailuresAreReturned(t *testing.T) {
	svc, ctx := setupService(t)
	created, err := svc.Create(ctx, adapters.CreateCustomerRequest{
		Name: "Ana Lima", Email: "ana@email.com", Type: "INDIVIDUAL",
		Document: "529.982.247-25", Phone: "11999990004",
	})
	require.NoError(t, err)

	broken := NewService(uow.NewInMemoryUoW(), listFailingRepo{svc.customerRepo}, svc.authRepo, &fakeOrders{})
	_, err = broken.GetByID(ctx, created.ID)
	assert.EqualError(t, err, "database unavailable")
	_, err = broken.GetByDocument(ctx, "529.982.247-25")
	assert.EqualError(t, err, "database unavailable")
	_, err = broken.Update(ctx, created.ID, adapters.UpdateCustomerRequest{Name: "Ana"})
	assert.EqualError(t, err, "database unavailable")
	assert.EqualError(t, broken.Delete(ctx, created.ID), "database unavailable")

	failing := NewService(failingUoW{}, svc.customerRepo, svc.authRepo, &fakeOrders{})
	_, err = failing.Update(ctx, created.ID, adapters.UpdateCustomerRequest{Name: "Ana", Email: "ana@email.com", Phone: "11999990004"})
	assert.EqualError(t, err, "unit of work unavailable")
}
