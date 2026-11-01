package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"northwind/callbackdirectory/config"
)

func TestFetchCountryEncodesCountryQuery(t *testing.T) {
	var receivedRawQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"name":"Puerto Rico","callingCodes":["1","1787","1939"]}]}`))
	}))
	defer server.Close()

	cfg := config.Config{
		UpstreamBaseURL: server.URL,
		RequestTimeout:  2 * time.Second,
	}
	c := NewHTTPCountriesClient(cfg)

	payload, err := c.FetchCountry("Puerto Rico")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedRawQuery != "name=Puerto+Rico" {
		t.Fatalf("expected encoded query name=Puerto+Rico, got %q", receivedRawQuery)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one country record, got %d", len(payload.Data))
	}
}
