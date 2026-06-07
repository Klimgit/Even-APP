package swagger

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var errSpecNotFound = errors.New("spec not found")

func loadSpec(client *http.Client, baseURL string) (*v3.Document, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/v1/openapi.yaml"
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s", errSpecNotFound, url)
		}
		return nil, fmt.Errorf("spec loading error — url: %s, status: %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return unmarshalSpec(data)
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}
