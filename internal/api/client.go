package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Ledgi Agent API.
type Client struct {
	APIKey  string
	BaseURL string // e.g. "https://api.ledgi.app"
	HTTP    *http.Client
}

// New creates a Client with sensible defaults.
func New(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// apiError is the standard error contract from the agent API.
type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Get performs a GET request and returns the raw JSON body.
func (c *Client) Get(path string, params map[string]string) (json.RawMessage, error) {
	u := c.BaseURL + "/api/v1" + path
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			if v != "" {
				q.Set(k, v)
			}
		}
		if encoded := q.Encode(); encoded != "" {
			u += "?" + encoded
		}
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body any) (json.RawMessage, error) {
	u := c.BaseURL + "/api/v1" + path
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest("POST", u, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	u := c.BaseURL + "/api/v1" + path
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if len(data) == 0 {
			return json.RawMessage("{}"), nil
		}
		return json.RawMessage(data), nil
	}

	return nil, mapError(resp.StatusCode, data)
}

func mapError(status int, body []byte) error {
	// Try to parse the standard error contract.
	var ae apiError
	if err := json.Unmarshal(body, &ae); err == nil && ae.Message != "" {
		return fmt.Errorf("%s", ae.Message)
	}

	// Try FastAPI detail format.
	var detail struct {
		Detail any `json:"detail"`
	}
	if err := json.Unmarshal(body, &detail); err == nil && detail.Detail != nil {
		switch v := detail.Detail.(type) {
		case string:
			return fmt.Errorf("%s", v)
		case map[string]any:
			if msg, ok := v["message"].(string); ok {
				return fmt.Errorf("%s", msg)
			}
		}
	}

	switch status {
	case 401:
		return fmt.Errorf("invalid or missing API key. Run: ledgi login --api-key <key>")
	case 403:
		return fmt.Errorf("API key lacks the required scope for this action")
	case 422:
		return fmt.Errorf("validation error: %s", string(body))
	case 429:
		return fmt.Errorf("rate limited. Try again later")
	default:
		if status >= 500 {
			return fmt.Errorf("server error (%d). Try again later", status)
		}
		return fmt.Errorf("unexpected error (%d): %s", status, string(body))
	}
}
