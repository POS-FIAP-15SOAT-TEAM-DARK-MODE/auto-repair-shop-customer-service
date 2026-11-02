// Package serviceorder is the HTTP client of the Service Order internal REST
// (contracts/internal-rest.md, section Service Order). It answers whether a customer or a
// vehicle has service orders, so that the delete operations can refuse to remove them.
package serviceorder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
)

const (
	existsPath          = "/internal/service-orders/exists"
	internalTokenHeader = "X-Internal-Token"
	defaultTimeout      = 3 * time.Second
	maxBodyBytes        = 4 << 10
)

// Client calls the Service Order internal REST. Calls are not retried: any failure is reported
// as app.ErrUpstreamUnavailable, and the caller must not change its state.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New builds a client for the Service Order at baseURL, authenticated with token
// (X-Internal-Token). A zero timeout means 3 seconds.
func New(baseURL, token string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// HasCustomerOrders reports whether the Service Order has orders for the customer.
func (c *Client) HasCustomerOrders(ctx context.Context, customerID string) (bool, error) {
	return c.exists(ctx, "customer_id", customerID)
}

// HasVehicleOrders reports whether the Service Order has orders for the vehicle.
func (c *Client) HasVehicleOrders(ctx context.Context, vehicleID string) (bool, error) {
	return c.exists(ctx, "vehicle_id", vehicleID)
}

func (c *Client) exists(ctx context.Context, param, value string) (bool, error) {
	query := url.Values{}
	query.Set(param, value)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+existsPath+"?"+query.Encode(), nil)
	if err != nil {
		return false, unavailable(err)
	}
	req.Header.Set(internalTokenHeader, c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, unavailable(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, unavailable(fmt.Errorf("unexpected status %d", resp.StatusCode))
	}

	var body struct {
		Exists *bool `json:"exists"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&body); err != nil {
		return false, unavailable(fmt.Errorf("invalid response body: %w", err))
	}
	if body.Exists == nil {
		return false, unavailable(fmt.Errorf("response has no exists field"))
	}
	return *body.Exists, nil
}

// unavailable wraps the cause for the server log. The cause is kept as text on purpose: the
// caller maps app.ErrUpstreamUnavailable to a fixed 502 body, and a single %w keeps the chain
// from being read as a joined (400) error.
func unavailable(cause error) error {
	return fmt.Errorf("%w: %s", app.ErrUpstreamUnavailable, cause.Error())
}
