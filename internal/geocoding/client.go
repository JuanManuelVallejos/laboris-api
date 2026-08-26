// Package geocoding is a thin client over the Google Maps Geocoding API — no
// SDK dependency, consistent with the project's minimal-dependencies style
// (see internal/storage for the same approach with Supabase).
package geocoding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{}}
}

var ErrNoResults = errors.New("no se encontró esa dirección")

type geocodeResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
}

// Geocode resolves a free-text address to coordinates. Returns ErrNoResults
// if Google couldn't match the address to anything.
func (c *Client) Geocode(ctx context.Context, address string) (lat, lng float64, err error) {
	endpoint := "https://maps.googleapis.com/maps/api/geocode/json?" + url.Values{
		"address": {address},
		"key":     {c.apiKey},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("geocoding request failed: %s", res.Status)
	}

	var parsed geocodeResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return 0, 0, err
	}

	if parsed.Status == "ZERO_RESULTS" || len(parsed.Results) == 0 {
		return 0, 0, ErrNoResults
	}
	if parsed.Status != "OK" {
		return 0, 0, fmt.Errorf("geocoding error: %s", parsed.Status)
	}

	loc := parsed.Results[0].Geometry.Location
	return loc.Lat, loc.Lng, nil
}
