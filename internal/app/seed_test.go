package app

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSeed_DisabledDoesNothing(t *testing.T) {
	t.Setenv("SEED_ENABLED", "false")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_SkipsWhenCustomersExist(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_CheckErrorStopsSeeding(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer`).WillReturnError(assert.AnError)
	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_InsertsStaffAndCustomers(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	t.Setenv("BCRYPT_COST", "4")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	// attendant + mechanic: user + role
	for i := 0; i < 2; i++ {
		mock.ExpectExec(`INSERT INTO "user"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO user_role`).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	// customer A: user + role + customer + vehicle
	mock.ExpectExec(`INSERT INTO "user"`).WithArgs("44444444-4444-4444-8444-000000000001", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO user_role`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO customer \(id, user_id, type, cpf, phone\)`).WithArgs("44444444-4444-4444-8444-000000000002", "44444444-4444-4444-8444-000000000001", "52998224725", "11999990001").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO vehicle`).WithArgs("44444444-4444-4444-8444-000000000003", "ABC-1234", "Toyota", "Corolla", 2020, "44444444-4444-4444-8444-000000000002").WillReturnResult(sqlmock.NewResult(1, 1))
	// customer B: user + role + customer + vehicle
	mock.ExpectExec(`INSERT INTO "user"`).WithArgs("44444444-4444-4444-8444-000000000011", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO user_role`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO customer \(id, user_id, type, cnpj, company_name, phone\)`).WithArgs("44444444-4444-4444-8444-000000000012", "44444444-4444-4444-8444-000000000011", "11222333000181", "Auto Transportes Silva", "11988880002").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO vehicle`).WithArgs("44444444-4444-4444-8444-000000000013", "XYZ1A23", "Ford", "Transit", 2022, "44444444-4444-4444-8444-000000000012").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_InsertErrorRollsBack(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	t.Setenv("BCRYPT_COST", "4")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user"`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_BeginErrorIsLogged(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin().WillReturnError(assert.AnError)

	RunSeed(context.Background(), db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnectWithDB(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ConnectWithDB(db)
	assert.Same(t, db, DBConnect())
}
