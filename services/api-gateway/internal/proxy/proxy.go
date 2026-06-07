package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Mount registers a reverse proxy for paths with the given prefix.
func Mount(mux *http.ServeMux, prefix, targetBase string) error {
	target, err := url.Parse(strings.TrimRight(targetBase, "/"))
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	mux.Handle(prefix, proxy)
	return nil
}

// MountPattern registers a reverse proxy for an exact Go 1.22 path pattern (may include {vars}).
func MountPattern(mux *http.ServeMux, pattern, targetBase string) error {
	target, err := url.Parse(strings.TrimRight(targetBase, "/"))
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	mux.Handle(pattern, proxy)
	return nil
}
