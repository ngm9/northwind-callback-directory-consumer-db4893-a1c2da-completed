package service

import (
	"testing"

	"northwind/callbackdirectory/client"
)

type stubClient struct {
	payload client.CountryPayload
}

func (s stubClient) FetchCountry(country string) (client.CountryPayload, error) {
	return s.payload, nil
}

func TestGetPhoneNumbersSamples(t *testing.T) {
	tests := []struct {
		name     string
		payload  client.CountryPayload
		country  string
		phone    string
		expected string
	}{
		{
			name: "single calling code",
			payload: client.CountryPayload{
				Data: []client.CountryRecord{{Name: "Afghanistan", CallingCodes: []string{"93"}}},
			},
			country:  "Afghanistan",
			phone:    "656445445",
			expected: "+93 656445445",
		},
		{
			name: "multiple calling codes uses highest index",
			payload: client.CountryPayload{
				Data: []client.CountryRecord{{Name: "Puerto Rico", CallingCodes: []string{"1", "1787", "1939"}}},
			},
			country:  "Puerto Rico",
			phone:    "564593986",
			expected: "+1939 564593986",
		},
		{
			name:     "unknown country returns minus one",
			payload:  client.CountryPayload{Data: []client.CountryRecord{}},
			country:  "Oceania",
			phone:    "987574876",
			expected: "-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := stubClient{payload: tc.payload}
			svc := NewDirectoryService(stub)
			result, err := svc.GetPhoneNumbers(tc.country, tc.phone)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}
