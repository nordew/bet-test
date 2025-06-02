package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nordew/bet-test/internal/model"
)

const (
	maxRetries    = 3
	retryDelaySec = 2
	url           = "https://jsonplaceholder.typicode.com/users"
)

type APIClient interface {
	FetchUsers(ctx context.Context) ([]model.User, error)
	SendUserToAPIB(ctx context.Context, payload model.UserPayload, apiBURL string) error
}

type apiClient struct {
	client *http.Client
}

func NewAPIClient() APIClient {
	return &apiClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *apiClient) FetchUsers(ctx context.Context) ([]model.User, error) {
	req, err := http.NewRequestWithContext(
		ctx, 
		http.MethodGet, 
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("API A: failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API A: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API A: non-OK status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("API A: failed to read response body: %w", err)
	}

	var users []model.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("API A: failed to unmarshal users: %w", err)
	}

	return users, nil
}

func (c *apiClient) SendUserToAPIB(
	ctx context.Context,
	payload model.UserPayload,
	apiBURL string,
) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("API B: failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			apiBURL,
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return fmt.Errorf("API B: failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt+1, err)
			select {
			case <-time.After(retryDelaySec * time.Second):
				continue
			case <-ctx.Done():
				return fmt.Errorf("API B: context cancelled during retry: %w", ctx.Err())
			}
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("API B: sent user %s (status %d)", payload.Email, resp.StatusCode)
			return nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf(
			"attempt %d: non-2xx status: %d, body: %s",
			attempt+1,
			resp.StatusCode,
			string(respBody),
		)

		select {
		case <-time.After(retryDelaySec * time.Second):
		case <-ctx.Done():
			return fmt.Errorf("API B: context cancelled during retry: %w", ctx.Err())
		}
	}

	log.Printf(
		"API B: failed to send user %s after %d attempts: %v",
		payload.Email, maxRetries, lastErr,
	)
	return fmt.Errorf("API B: failed after %d attempts: %w", maxRetries, lastErr)
} 