package ready

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var checkOrder = []string{"auth", "media", "lexicon", "content", "learning"}

// CheckBackends GETs /api/v1/ready on each upstream. Returns on first failure.
func CheckBackends(ctx context.Context, backends map[string]string) error {
	client := &http.Client{Timeout: 3 * time.Second}

	for _, name := range checkOrder {
		base, ok := backends[name]
		if !ok || base == "" {
			return fmt.Errorf("%s: URL not configured", name)
		}
		url := strings.TrimRight(base, "/") + "/api/v1/ready"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("%s: build request: %w", name, err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: ready returned %d", name, resp.StatusCode)
		}
	}
	return nil
}
