package serviceorder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testToken = "test-internal-token"

// newServer answers every request with the given status and body, and checks the request
// contract (path, query parameter, X-Internal-Token) through check when it is not nil.
func newServer(t *testing.T, check func(r *http.Request), status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestHasCustomerOrders_ExistsTrueAndFalse(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"exists true", `{"exists":true}`, true},
		{"exists false", `{"exists":false}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, func(r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/internal/service-orders/exists", r.URL.Path)
				assert.Equal(t, "cust-1", r.URL.Query().Get("customer_id"))
				assert.Empty(t, r.URL.Query().Get("vehicle_id"))
			}, http.StatusOK, tc.body)

			got, err := New(srv.URL, testToken, 0).HasCustomerOrders(context.Background(), "cust-1")
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHasVehicleOrders_UsesVehicleQueryParam(t *testing.T) {
	srv := newServer(t, func(r *http.Request) {
		assert.Equal(t, "veh-9", r.URL.Query().Get("vehicle_id"))
		assert.Empty(t, r.URL.Query().Get("customer_id"))
	}, http.StatusOK, `{"exists":true}`)

	got, err := New(srv.URL, testToken, 0).HasVehicleOrders(context.Background(), "veh-9")
	require.NoError(t, err)
	assert.True(t, got)
}

func TestExists_SendsInternalToken(t *testing.T) {
	srv := newServer(t, func(r *http.Request) {
		assert.Equal(t, testToken, r.Header.Get("X-Internal-Token"))
	}, http.StatusOK, `{"exists":false}`)

	_, err := New(srv.URL, testToken, 0).HasCustomerOrders(context.Background(), "cust-1")
	require.NoError(t, err)
}

func TestExists_TrailingSlashInBaseURL(t *testing.T) {
	srv := newServer(t, func(r *http.Request) {
		assert.Equal(t, "/internal/service-orders/exists", r.URL.Path)
	}, http.StatusOK, `{"exists":false}`)

	_, err := New(srv.URL+"/", testToken, 0).HasCustomerOrders(context.Background(), "cust-1")
	require.NoError(t, err)
}

func TestExists_UpstreamFailuresAreUnavailable(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"5xx", http.StatusInternalServerError, `{"code":500,"error":"boom"}`},
		{"503", http.StatusServiceUnavailable, ""},
		{"401 wrong token", http.StatusUnauthorized, `{"code":401,"error":"Unauthorized"}`},
		{"422 bad request", http.StatusUnprocessableEntity, `{"code":422,"error":"invalid data"}`},
		{"malformed json", http.StatusOK, `{"exists":`},
		{"missing exists", http.StatusOK, `{}`},
		{"null exists", http.StatusOK, `{"exists":null}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, nil, tc.status, tc.body)

			got, err := New(srv.URL, testToken, 0).HasCustomerOrders(context.Background(), "cust-1")
			require.ErrorIs(t, err, app.ErrUpstreamUnavailable)
			assert.False(t, got)
		})
	}
}

func TestExists_TimeoutIsUnavailable(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() {
		close(release)
		srv.Close()
	})

	client := New(srv.URL, testToken, 50*time.Millisecond)
	_, err := client.HasCustomerOrders(context.Background(), "cust-1")
	require.ErrorIs(t, err, app.ErrUpstreamUnavailable)
}

func TestExists_ConnectionRefusedIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	_, err := New(url, testToken, 0).HasVehicleOrders(context.Background(), "veh-1")
	require.ErrorIs(t, err, app.ErrUpstreamUnavailable)
}

func TestExists_InvalidURLIsUnavailable(t *testing.T) {
	_, err := New("http://[::1", testToken, 0).HasCustomerOrders(context.Background(), "cust-1")
	require.ErrorIs(t, err, app.ErrUpstreamUnavailable)
}

func TestExists_CanceledContextIsUnavailable(t *testing.T) {
	srv := newServer(t, nil, http.StatusOK, `{"exists":false}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(srv.URL, testToken, 0).HasCustomerOrders(ctx, "cust-1")
	require.ErrorIs(t, err, app.ErrUpstreamUnavailable)
}

func TestNew_DefaultTimeoutIsThreeSeconds(t *testing.T) {
	assert.Equal(t, 3*time.Second, New("http://x", testToken, 0).httpClient.Timeout)
	assert.Equal(t, 5*time.Second, New("http://x", testToken, 5*time.Second).httpClient.Timeout)
}
