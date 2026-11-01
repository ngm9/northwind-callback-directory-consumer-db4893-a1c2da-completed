package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"northwind/callbackdirectory/client"
	"northwind/callbackdirectory/service"
)

type stubClient struct {
	payload client.CountryPayload
	err     error
}

func (s stubClient) FetchCountry(country string) (client.CountryPayload, error) {
	if s.err != nil {
		return client.CountryPayload{}, s.err
	}
	return s.payload, nil
}

func TestEndpointSamples(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		payload      client.CountryPayload
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name:  "single calling code",
			query: "/phone-numbers?country=Afghanistan&phone=656445445",
			payload: client.CountryPayload{
				Data: []client.CountryRecord{{Name: "Afghanistan", CallingCodes: []string{"93"}}},
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"phone": "+93 656445445"},
		},
		{
			name:  "multiple calling codes uses highest index",
			query: "/phone-numbers?country=Puerto%20Rico&phone=564593986",
			payload: client.CountryPayload{
				Data: []client.CountryRecord{{Name: "Puerto Rico", CallingCodes: []string{"1", "1787", "1939"}}},
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"phone": "+1939 564593986"},
		},
		{
			name:         "unknown country returns minus one",
			query:        "/phone-numbers?country=Oceania&phone=987574876",
			payload:      client.CountryPayload{Data: []client.CountryRecord{}},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"phone": "-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := stubClient{payload: tc.payload}
			svc := service.NewDirectoryService(stub)
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)

			if rec.Code != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, rec.Code)
			}
			assertJSONBody(t, rec, tc.expectedBody)
		})
	}
}

func TestEndpointValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectedBody map[string]string
	}{
		{
			name:         "missing country",
			query:        "/phone-numbers?phone=656445445",
			expectedBody: map[string]string{"error": "'country' is required and cannot be blank."},
		},
		{
			name:         "blank country",
			query:        "/phone-numbers?country=%20%20%20&phone=656445445",
			expectedBody: map[string]string{"error": "'country' is required and cannot be blank."},
		},
		{
			name:         "missing phone",
			query:        "/phone-numbers?country=Afghanistan",
			expectedBody: map[string]string{"error": "'phone' is required and cannot be blank."},
		},
		{
			name:         "blank phone",
			query:        "/phone-numbers?country=Afghanistan&phone=%20%20%20",
			expectedBody: map[string]string{"error": "'phone' is required and cannot be blank."},
		},
	}

	stub := stubClient{payload: client.CountryPayload{Data: []client.CountryRecord{}}}
	svc := service.NewDirectoryService(stub)
	h := NewHandler(svc)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
			assertJSONBody(t, rec, tc.expectedBody)
		})
	}
}

func TestEndpointUpstreamErrorReturnsInternalServerErrorJSON(t *testing.T) {
	stub := stubClient{err: errors.New("upstream timeout")}
	svc := service.NewDirectoryService(stub)
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/phone-numbers?country=Afghanistan&phone=656445445", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	assertJSONBody(t, rec, map[string]string{"error": "failed to resolve callback number."})
}

func assertJSONBody(t *testing.T, rec *httptest.ResponseRecorder, expected map[string]string) {
	t.Helper()
	var actual map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}
	for key, expectedValue := range expected {
		actualValue := actual[key]
		if actualValue != expectedValue {
			t.Fatalf("expected body[%s]=%q, got %q", key, expectedValue, actualValue)
		}
	}
}
