package repository

import (
	"context"
	"database/sql"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"go.uber.org/zap"
)

const (
	upsertCustomerQuery = `
		INSERT INTO customer (id, user_id, type, cpf, cnpj, company_name, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			phone        = EXCLUDED.phone,
			company_name = EXCLUDED.company_name,
			updated_at   = NOW()
	`

	deleteCustomerQuery = `DELETE FROM customer WHERE id = $1`

	listCustomerBaseQuery = `
		SELECT c.id, c.user_id, c.type,
		       COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
		       u.name, u.email
		FROM customer c
		JOIN "user" u ON u.id = c.user_id`
)

type postgresRepository struct{}

func NewPostgres() interfaces.CustomerRepository {
	return &postgresRepository{}
}

func (r *postgresRepository) Save(ctx context.Context, c *domain.Customer) error {
	tx, err := uowPkg.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, upsertCustomerQuery,
		c.ID,
		c.UserID,
		string(c.Type),
		nullableString(c.CPF),
		nullableString(c.CNPJ),
		nullableString(c.CompanyName),
		c.Phone,
	)
	if err != nil {
		mapped := pgPkg.Error(ctx, err)
		logger.Of(ctx).Warn("customer.save_failed",
			zap.String("operation", "save_customer"),
			zap.String("entity_id", c.ID),
			zap.Error(err),
		)
		return mapped
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	tx, err := uowPkg.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, deleteCustomerQuery, id); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *postgresRepository) List(ctx context.Context, params adapters.ListCustomerParams) ([]domain.Customer, error) {
	db, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	query, args := buildListQuery(params)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var customers []domain.Customer
	for rows.Next() {
		c, scanErr := scanCustomer(rows)
		if scanErr != nil {
			return nil, pgPkg.Error(ctx, scanErr)
		}
		customers = append(customers, c)
	}

	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return customers, nil
}

func buildListQuery(params adapters.ListCustomerParams) (string, []any) {
	switch {
	case params.ID != "":
		return listCustomerBaseQuery + ` WHERE c.id = $1`, []any{params.ID}
	case params.Document != "":
		return listCustomerBaseQuery + ` WHERE c.cpf = $1 OR c.cnpj = $1`, []any{params.Document}
	case params.UserID != "":
		return listCustomerBaseQuery + ` WHERE c.user_id = $1`, []any{params.UserID}
	default:
		return listCustomerBaseQuery, nil
	}
}

func scanCustomer(rows *sql.Rows) (domain.Customer, error) {
	var c domain.Customer
	user := &authDomain.User{}

	err := rows.Scan(
		&c.ID, &c.UserID, &c.Type,
		&c.CPF, &c.CNPJ, &c.CompanyName, &c.Phone,
		&user.Name, &user.Email,
	)
	if err != nil {
		return domain.Customer{}, err
	}

	user.ID = c.UserID
	c.User = user
	return c, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Ensure compile-time interface satisfaction.
var _ interfaces.CustomerRepository = (*postgresRepository)(nil)
