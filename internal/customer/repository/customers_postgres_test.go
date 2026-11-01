package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	app.ConnectWithDB(db)
	return db, mock
}

func TestPostgres_SaveInsertsCustomerInTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPostgres()
	c := &domain.Customer{ID: "c1", UserID: "u1", Type: domain.CompanyCustomerType, CNPJ: "45997418000153", CompanyName: "Ltda", Phone: "1"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").
		WithArgs("c1", "u1", "COMPANY", nil, "45997418000153", "Ltda", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, c)
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_SaveWithoutTransactionFails(t *testing.T) {
	newMockDB(t)
	err := NewPostgres().Save(context.Background(), &domain.Customer{ID: "c1"})
	assert.Error(t, err)
}

func TestPostgres_SaveMapsDuplicateToConflict(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").WillReturnError(errors.New("pq: duplicate key value violates unique constraint"))
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return NewPostgres().Save(ctx, &domain.Customer{ID: "c1", Phone: "1"})
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, app.ErrDataConflict)
}

func TestPostgres_DeleteRunsDeleteStatement(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM customer WHERE id = \$1`).WithArgs("c1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return NewPostgres().Delete(ctx, "c1")
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_DeleteWithoutTransactionFails(t *testing.T) {
	newMockDB(t)
	assert.Error(t, NewPostgres().Delete(context.Background(), "c1"))
}

func TestPostgres_ListByFilters(t *testing.T) {
	_, mock := newMockDB(t)
	repo := NewPostgres()
	cols := []string{"id", "user_id", "type", "cpf", "cnpj", "company_name", "phone", "name", "email"}

	mock.ExpectQuery(`WHERE c.id = \$1`).WithArgs("c1").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("c1", "u1", "INDIVIDUAL", "52998224725", "", "", "1", "João", "j@s.com"))
	got, err := repo.List(context.Background(), adapters.ListCustomerParams{ID: "c1"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "52998224725", got[0].CPF)
	assert.Equal(t, "João", got[0].User.Name)
	assert.Equal(t, "u1", got[0].User.ID)

	mock.ExpectQuery(`WHERE c.cpf = \$1 OR c.cnpj = \$1`).WithArgs("doc").
		WillReturnRows(sqlmock.NewRows(cols))
	got, err = repo.List(context.Background(), adapters.ListCustomerParams{Document: "doc"})
	require.NoError(t, err)
	assert.Empty(t, got)

	mock.ExpectQuery(`WHERE c.user_id = \$1`).WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(cols))
	_, err = repo.List(context.Background(), adapters.ListCustomerParams{UserID: "u1"})
	require.NoError(t, err)

	mock.ExpectQuery(`FROM customer c`).WillReturnRows(sqlmock.NewRows(cols))
	_, err = repo.List(context.Background(), adapters.ListCustomerParams{})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_ListQueryErrorIsMapped(t *testing.T) {
	_, mock := newMockDB(t)
	mock.ExpectQuery(`FROM customer c`).WillReturnError(errors.New("connection reset"))
	_, err := NewPostgres().List(context.Background(), adapters.ListCustomerParams{})
	assert.Error(t, err)
}

func TestBuildListQuery(t *testing.T) {
	q, args := buildListQuery(adapters.ListCustomerParams{ID: "x"})
	assert.Contains(t, q, "WHERE c.id = $1")
	assert.Equal(t, []any{"x"}, args)

	q, args = buildListQuery(adapters.ListCustomerParams{})
	assert.NotContains(t, q, "WHERE")
	assert.Nil(t, args)
}

func TestNullableString(t *testing.T) {
	assert.Nil(t, nullableString(""))
	assert.Equal(t, "x", nullableString("x"))
}
