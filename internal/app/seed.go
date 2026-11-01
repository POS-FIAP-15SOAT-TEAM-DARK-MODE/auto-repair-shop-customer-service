package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const checkSeededQuery = `SELECT COUNT(*) FROM customer`

// Fixed IDs (docs/fase4/contracts/resources.md): e2e tests and fixtures depend on them.
const (
	attendantUserID = "33333333-3333-4333-8333-000000000001"
	mechanicUserID  = "33333333-3333-4333-8333-000000000002"

	customerAUserID    = "44444444-4444-4444-8444-000000000001"
	customerACustomer  = "44444444-4444-4444-8444-000000000002"
	customerAVehicle   = "44444444-4444-4444-8444-000000000003"
	customerBUserID    = "44444444-4444-4444-8444-000000000011"
	customerBCustomer  = "44444444-4444-4444-8444-000000000012"
	customerBVehicle   = "44444444-4444-4444-8444-000000000013"
	userRoleIDTemplate = "55555555-5555-4555-8555-%012d"
)

// RunSeed inserts the staff users, the customers A/B and their vehicles, only when
// SEED_ENABLED=true and the customer table is empty.
func RunSeed(ctx context.Context, db *sql.DB) {
	log := logger.Global()

	if env.GetString("SEED_ENABLED", "false") != "true" {
		log.Info("seed: disabled (SEED_ENABLED!=true)")
		return
	}

	seeded, err := isAlreadySeeded(ctx, db)
	if err != nil {
		log.Error(fmt.Errorf("seed: failed to check seed state: %w", err))
		return
	}
	if seeded {
		log.Info("seed: data already present, skipping")
		return
	}

	log.Info("seed: starting database seeding")

	if err := runInTransaction(ctx, db, insertSeedData); err != nil {
		log.Error(fmt.Errorf("seed: failed to insert seed data: %w", err))
		return
	}

	log.Info("seed: database seeded successfully")
}

func isAlreadySeeded(ctx context.Context, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, checkSeededQuery).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func runInTransaction(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func insertSeedData(ctx context.Context, tx *sql.Tx) error {
	log := logger.Global()
	bcryptCost := env.GetInt("BCRYPT_COST", 12)
	roleN := 0
	nextRoleID := func() string {
		roleN++
		return fmt.Sprintf(userRoleIDTemplate, roleN)
	}

	if err := insertStaffUser(ctx, tx, nextRoleID(), attendantUserID, "attendant", "attendant@autorepairshop.com", "attendant123", "ATTENDANT", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: attendant user inserted", zap.String("user_id", attendantUserID))

	if err := insertStaffUser(ctx, tx, nextRoleID(), mechanicUserID, "mechanic", "mechanic@autorepairshop.com", "mechanic123", "MECHANIC", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: mechanic user inserted", zap.String("user_id", mechanicUserID))

	customerACPF := "529.982.247-25"
	if err := insertIndividualCustomer(ctx, tx, nextRoleID(), customerAUserID, customerACustomer, "João Silva", "joao.silva@email.com", customerACPF, "11999990001", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: individual customer inserted", zap.String("customer_id", customerACustomer))

	if err := insertVehicle(ctx, tx, customerAVehicle, "ABC-1234", "Toyota", "Corolla", 2020, customerACustomer); err != nil {
		return err
	}
	log.Info("seed: vehicle inserted", zap.String("vehicle_id", customerAVehicle))

	customerBCNPJ := "11.222.333/0001-81"
	if err := insertCompanyCustomer(ctx, tx, nextRoleID(), customerBUserID, customerBCustomer, "Transportes Silva Ltda", "contato@transportessilva.com.br", customerBCNPJ, "Auto Transportes Silva", "11988880002", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: company customer inserted", zap.String("customer_id", customerBCustomer))

	if err := insertVehicle(ctx, tx, customerBVehicle, "XYZ1A23", "Ford", "Transit", 2022, customerBCustomer); err != nil {
		return err
	}
	log.Info("seed: vehicle inserted", zap.String("vehicle_id", customerBVehicle))

	return nil
}

// --- Staff users (ATTENDANT / MECHANIC) ---

func insertStaffUser(ctx context.Context, tx *sql.Tx, userRoleID, userID, name, email, rawPassword, roleName string, bcryptCost int) error {
	hash, err := hashPassword(rawPassword, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert staff user %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, userRoleID, userID, roleName); err != nil {
		return fmt.Errorf("assign role %q to user %q: %w", roleName, email, err)
	}

	return nil
}

// --- Individual customer (CPF) ---

func insertIndividualCustomer(ctx context.Context, tx *sql.Tx, userRoleID, userID, customerID, name, email, cpf, phone string, bcryptCost int) error {
	hash, err := hashPassword(cpf, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for customer %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert user for individual customer %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, userRoleID, userID, "CUSTOMER"); err != nil {
		return fmt.Errorf("assign CUSTOMER role to %q: %w", email, err)
	}

	const insertCustomer = `INSERT INTO customer (id, user_id, type, cpf, phone) VALUES ($1, $2, 'INDIVIDUAL', $3, $4)`
	if _, err := tx.ExecContext(ctx, insertCustomer, customerID, userID, cpf, phone); err != nil {
		return fmt.Errorf("insert individual customer %q: %w", email, err)
	}

	return nil
}

// --- Company customer (CNPJ) ---

func insertCompanyCustomer(ctx context.Context, tx *sql.Tx, userRoleID, userID, customerID, name, email, cnpj, companyName, phone string, bcryptCost int) error {
	hash, err := hashPassword(cnpj, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for company customer %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert user for company customer %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, userRoleID, userID, "CUSTOMER"); err != nil {
		return fmt.Errorf("assign CUSTOMER role to %q: %w", email, err)
	}

	const insertCustomer = `INSERT INTO customer (id, user_id, type, cnpj, company_name, phone) VALUES ($1, $2, 'COMPANY', $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, insertCustomer, customerID, userID, cnpj, companyName, phone); err != nil {
		return fmt.Errorf("insert company customer %q: %w", email, err)
	}

	return nil
}

// --- Vehicle ---

func insertVehicle(ctx context.Context, tx *sql.Tx, vehicleID, plate, brand, model string, year int, customerID string) error {
	const q = `INSERT INTO vehicle (id, license_plate, brand, model, year, customer_id) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.ExecContext(ctx, q, vehicleID, plate, brand, model, year, customerID); err != nil {
		return fmt.Errorf("insert vehicle %q: %w", plate, err)
	}
	return nil
}

// --- Helpers ---

func hashPassword(raw string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
