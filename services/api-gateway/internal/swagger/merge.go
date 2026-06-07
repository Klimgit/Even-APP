package swagger

import (
	"errors"
	"fmt"

	"github.com/pb33f/libopenapi"
)

var backendOrder = []string{"auth", "media", "lexicon", "content", "learning"}

// FetchMerged downloads openapi.yaml from each backend and merges paths/components via libopenapi.
func FetchMerged(backends map[string]string) ([]byte, error) {
	defer libopenapi.ClearAllCaches()

	sb := &specBuilder{}
	if err := sb.init(); err != nil {
		return nil, err
	}

	client := newHTTPClient()
	for _, name := range backendOrder {
		base, ok := backends[name]
		if !ok || base == "" {
			continue
		}
		doc, err := loadSpec(client, base)
		if err != nil {
			if errors.Is(err, errSpecNotFound) {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if err := sb.add(doc); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}

	return sb.build().Render()
}
