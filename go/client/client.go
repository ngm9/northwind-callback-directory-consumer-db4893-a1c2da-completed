package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"northwind/callbackdirectory/config"
)

type CountryRecord struct {
	Name         string   `json:"name"`
	CallingCodes []string `json:"callingCodes"`
}

type CountryPayload struct {
	Data []CountryRecord `json:"data"`
}

type CountriesClient interface {
	FetchCountry(country string) (CountryPayload, error)
}

type HTTPCountriesClient struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewHTTPCountriesClient(cfg config.Config) *HTTPCountriesClient {
	return &HTTPCountriesClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (c *HTTPCountriesClient) FetchCountry(country string) (CountryPayload, error) {
	encodedCountry := url.QueryEscape(country)
	requestURL := fmt.Sprintf("%s?name=%s", c.cfg.UpstreamBaseURL, encodedCountry)
	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return CountryPayload{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return CountryPayload{}, fmt.Errorf("upstream countries API request failed with status %d", resp.StatusCode)
	}

	var payload CountryPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return CountryPayload{}, err
	}
	return payload, nil
}
