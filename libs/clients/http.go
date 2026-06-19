package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const InternalTokenHeader = "X-Internal-Token"

// HTTP is a minimal service-to-service JSON client.
type HTTP struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewHTTP(baseURL, token string) *HTTP {
	return &HTTP{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *HTTP) Available() bool {
	return c != nil && c.BaseURL != ""
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("upstream HTTP %d: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("upstream HTTP %d", e.Status)
}

func (c *HTTP) DoJSON(ctx context.Context, method, path string, in any, out any) error {
	if !c.Available() {
		return fmt.Errorf("service URL not configured")
	}
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set(InternalTokenHeader, c.Token)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(raw))
		var er struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		_ = json.Unmarshal(raw, &er)
		if er.Message != "" {
			msg = er.Message
		} else if er.Error != "" {
			msg = er.Error
		}
		return &APIError{Status: resp.StatusCode, Message: msg}
	}
	if out == nil || len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, out)
}
