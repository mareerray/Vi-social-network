package utils

import (
	"fmt"
	"net/http"
	"strings"
)

// AbsURL returns an absolute URL for the given path. If the input is already
// an absolute URL (starts with http:// or https://) it is returned untouched.
// Otherwise the function will prefix the request's scheme and host.
func AbsURL(r *http.Request, path string) string {
	if path == "" {
		return ""
	}
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	// Already absolute
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	scheme := "http"
	if r != nil {
		if r.TLS != nil {
			scheme = "https"
		} else if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
	}
	host := "localhost"
	if r != nil && r.Host != "" {
		host = r.Host
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, p)
}
