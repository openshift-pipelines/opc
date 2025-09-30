// Copyright 2025 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents the Tekton Assistant API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Config holds configuration for the client
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new Tekton Assistant API client
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// ExplainFailureRequest represents the request parameters for explaining TaskRun failures
type ExplainFailureRequest struct {
	Namespace string
	TaskRun   string
}

// ExplainPipelineRunFailureRequest represents the request parameters for explaining PipelineRun failures
type ExplainPipelineRunFailureRequest struct {
	Namespace   string
	PipelineRun string
}

// ExplainFailure calls the /taskrun/explainFailure endpoint to get AI-powered diagnosis
func (c *Client) ExplainFailure(ctx context.Context, req *ExplainFailureRequest) (string, error) {
	if req.Namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if req.TaskRun == "" {
		return "", fmt.Errorf("taskrun name is required")
	}

	// Construct the URL with query parameters
	endpoint := fmt.Sprintf("%s/taskrun/explainFailure", c.baseURL)
	params := url.Values{}
	params.Add("namespace", req.Namespace)
	params.Add("name", req.TaskRun)

	requestURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "tkn-assist/dev")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// ExplainPipelineRunFailure calls the /pipelinerun/explainFailure endpoint to get AI-powered diagnosis
func (c *Client) ExplainPipelineRunFailure(ctx context.Context, req *ExplainPipelineRunFailureRequest) (string, error) {
	if req.Namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if req.PipelineRun == "" {
		return "", fmt.Errorf("pipelinerun name is required")
	}

	// Construct the URL with query parameters
	endpoint := fmt.Sprintf("%s/pipelinerun/explainFailure", c.baseURL)
	params := url.Values{}
	params.Add("namespace", req.Namespace)
	params.Add("name", req.PipelineRun)

	requestURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "tkn-assist/dev")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// Health checks if the API server is healthy
func (c *Client) Health(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
