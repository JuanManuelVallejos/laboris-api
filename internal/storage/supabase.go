// Package storage is a thin client over the Supabase Storage REST API — no
// AWS SDK dependency, consistent with the project's minimal-dependencies
// style (pgx used raw, no ORM). Supabase Storage speaks the real S3
// protocol under the hood and exposes an S3-compatible endpoint, so a
// future move to plain AWS S3 is a matter of swapping the endpoint and
// credentials, not rewriting the upload logic that calls this package.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SupabaseClient struct {
	baseURL        string
	serviceRoleKey string
	bucket         string
	httpClient     *http.Client
}

func NewSupabaseClient(baseURL, serviceRoleKey, bucket string) *SupabaseClient {
	return &SupabaseClient{
		baseURL:        baseURL,
		serviceRoleKey: serviceRoleKey,
		bucket:         bucket,
		httpClient:     &http.Client{},
	}
}

// Upload stores contents at path within the configured bucket. The bucket is
// expected to be private — read access is granted per request via SignedURL,
// not via a permanent public URL.
func (c *SupabaseClient) Upload(ctx context.Context, path string, contents io.Reader, contentType string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, contents)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("supabase storage upload failed: %s: %s", res.Status, string(body))
	}
	return nil
}

// SignedURL requests a time-limited signed URL for reading the object at
// path, valid for ttl. Meant for private buckets: nobody can read the file
// without a URL our backend explicitly issued and that expires on its own.
func (c *SupabaseClient) SignedURL(ctx context.Context, path string, ttl time.Duration) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", c.baseURL, c.bucket, path)
	body, err := json.Marshal(map[string]int{"expiresIn": int(ttl.Seconds())})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		respBody, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("supabase storage sign failed: %s: %s", res.Status, string(respBody))
	}

	var parsed struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return c.baseURL + "/storage/v1" + parsed.SignedURL, nil
}

// Delete removes the object at path from the configured bucket.
func (c *SupabaseClient) Delete(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("supabase storage delete failed: %s: %s", res.Status, string(body))
	}
	return nil
}
