package bootstrap

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestHttpContainer(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	c := httpContainer(db)

	assert.NotNil(t, c)
	assert.NotNil(t, c.UserHandler)
	assert.NotNil(t, c.CustomerHandler)
	assert.NotNil(t, c.VehicleHandler)
	assert.NotNil(t, c.CustomerInternalHandler)
	assert.NotNil(t, c.VehicleInternalHandler)
}

func TestMiddlewaresContainer(t *testing.T) {
	m := middlewaresContainer()
	assert.NotNil(t, m)
	assert.Contains(t, *m, "Logger")
	assert.Contains(t, *m, "Recovery")
	assert.Contains(t, *m, "ErrorHandler")
}

func TestHTTPServerBootstrap(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	app.ConnectWithDB(db)

	s := HTTPServer()
	assert.NotNil(t, s)
}
