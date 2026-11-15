package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	repository "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	app.ConnectWithDB(db)
	repo := repository.NewPostgres(time.Minute)

	email := "admin@example.com"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash"}).
				AddRow("u1", "Admin", email, "hashed"))

		u, err := repo.GetByEmail(context.Background(), email)
		assert.NoError(t, err)
		assert.Equal(t, "u1", u.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByEmail(context.Background(), email)
		assert.ErrorIs(t, err, authDomain.ErrInvalidUserCredentials)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnError(assert.AnError)

		_, err := repo.GetByEmail(context.Background(), email)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "user"`).WithArgs("New Name", "new@e.com", "u1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "u1", "New Name", "new@e.com")
		})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "user"`).WithArgs("New Name", "new@e.com", "u1").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "u1", "New Name", "new@e.com")
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_SaveUserRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "MECHANIC").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.SaveUserRole(ctx, "u1", authDomain.MECHANIC)
		})
		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.SaveUserRole(ctx, "u1", authDomain.MECHANIC)
		})
		assert.Error(t, err)
	})

	t.Run("insert error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "MECHANIC").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.SaveUserRole(ctx, "u1", authDomain.MECHANIC)
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`DELETE FROM "user"`).WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, "u1")
		})
		assert.NoError(t, err)
	})

	t.Run("roles delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, "u1")
		})
		assert.Error(t, err)
	})

	t.Run("user delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`DELETE FROM "user"`).WithArgs("u1").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, "u1")
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user"`).
		WithArgs(u.ID, u.Name, u.Email, u.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, u)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, u)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Create_MissingTx(t *testing.T) {
	repo := repository.NewPostgres(time.Minute)
	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	err := repo.Save(context.Background(), u)
	assert.Error(t, err)
}
