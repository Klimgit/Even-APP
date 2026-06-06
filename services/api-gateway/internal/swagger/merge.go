package swagger

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FetchMerged downloads openapi.yaml from each backend and concatenates as comments.
// Skeleton implementation — full merge via libopenapi in a later phase.
func FetchMerged(backends map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	var b strings.Builder
	b.WriteString("openapi: 3.0.3\ninfo:\n  title: even-app (merged skeleton)\n  version: 0.1.0\n")
	b.WriteString("paths:\n  /api/v1/health:\n    get:\n      summary: Gateway health\n      responses:\n        '200':\n          description: OK\n")
	b.WriteString("# --- backend specs (skeleton, not merged paths yet) ---\n")

	for name, base := range backends {
		u := strings.TrimRight(base, "/") + "/api/v1/openapi.yaml"
		resp, err := client.Get(u)
		if err != nil {
			b.WriteString(fmt.Sprintf("# %s: fetch error: %v\n", name, err))
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		b.WriteString(fmt.Sprintf("# --- %s ---\n", name))
		b.Write(body)
		b.WriteString("\n")
	}
	return []byte(b.String()), nil
}
