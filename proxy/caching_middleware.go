package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	cpCache "caching-proxy/cache"
)

// CachedResponse represents a cached version of an HTTP response.
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// ToHTTPResponse converts a CachedResponse to an *http.Response.
func (cr *CachedResponse) ToHTTPResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: cr.StatusCode,
		Header:     cr.Header.Clone(),
		Body:       io.NopCloser(bytes.NewReader(cr.Body)),
		Request:    req,
	}
}

// CachingTransport is a custom http.RoundTripper that caches responses and follows redirects.
type CachingTransport struct {
	Transport http.RoundTripper
}

// RoundTrip checks the cache for a response before forwarding the request. It follows redirect
// responses automatically (up to a limit) so that the final response (which is returned to the client)
// always has our custom X-Cache header.
func (t *CachingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Save the original request key based on the initial URL.
	originalKey := req.Method + ":" + req.URL.String()
	log.Printf("[CachingTransport] Processing request: %s", originalKey)

	// If the request contains a bypass flag, skip caching.
	if strings.Contains(req.URL.RawQuery, "bypass-cache=true") {
		log.Printf("[CachingTransport] Cache bypass requested for %s", originalKey)
		goto performRequest
	}

	// Check for cached response.
	if cached, found := cpCache.InMemoryCache.Get(originalKey); found {
		if cachedResp, ok := cached.(*CachedResponse); ok {
			log.Printf("[CachingTransport] Cache HIT for %s", originalKey)
			resp := cachedResp.ToHTTPResponse(req)
			resp.Header.Set("X-Cache", "HIT")
			return resp, nil
		}
	}

performRequest:
	log.Printf("[CachingTransport] Cache MISS for %s – forwarding request to origin", originalKey)
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		log.Printf("[CachingTransport] Error forwarding request for %s: %v", originalKey, err)
		return nil, err
	}

	// --- Begin Redirect Following Logic ---
	// If the origin returns a redirect, follow the chain here (up to 10 times).
	redirectCount := 0
	for resp.StatusCode >= 300 && resp.StatusCode < 400 && redirectCount < 10 {
		location := resp.Header.Get("Location")
		if location == "" {
			break
		}
		log.Printf("[CachingTransport] Received redirect (status %d) for %s: Location=%s", resp.StatusCode, originalKey, location)

		// Resolve the redirect location relative to the current request URL.
		newURL, err := req.URL.Parse(location)
		if err != nil {
			log.Printf("[CachingTransport] Error parsing redirect URL: %v", err)
			break
		}

		// For a 303 See Other, force method to GET and drop the body.
		if resp.StatusCode == http.StatusSeeOther {
			req.Method = http.MethodGet
			req.Body = nil
		}

		// Update the request URL to the new location.
		req.URL = newURL
		log.Printf("[CachingTransport] Following redirect to: %s", newURL.String())

		// Make a new RoundTrip call with the updated URL.
		resp, err = t.Transport.RoundTrip(req)
		if err != nil {
			log.Printf("[CachingTransport] Error following redirect for %s: %v", originalKey, err)
			return nil, err
		}
		redirectCount++
	}
	if redirectCount >= 10 {
		log.Printf("[CachingTransport] Maximum redirects reached for %s", originalKey)
	}
	// --- End Redirect Following Logic ---

	// Now we have the final response. Cache it if appropriate.
	if isCacheable(resp.Header) {
		log.Printf("[CachingTransport] Caching final response for %s", originalKey)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[CachingTransport] Error reading response body for %s: %v", originalKey, err)
			return nil, err
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("[CachingTransport] Error closing response body for %s: %v", originalKey, err)
			return nil, err
		}
		cachedResp := &CachedResponse{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       bodyBytes,
		}
		cpCache.InMemoryCache.Set(originalKey, cachedResp, 5*time.Minute)
		// Replace the body for downstream read.
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		log.Printf("[CachingTransport] Final response for %s not cacheable. Cache-Control: %s", originalKey, resp.Header.Get("Cache-Control"))
	}

	// Mark the final response as missed.
	resp.Header.Set("X-Cache", "MISS")
	log.Printf("[CachingTransport] Returning final response for %s with X-Cache=%s", originalKey, resp.Header.Get("X-Cache"))
	return resp, nil
}

// isCacheable checks whether the response should be cached by inspecting Cache-Control.
func isCacheable(header http.Header) bool {
	cacheControl := header.Get("Cache-Control")
	if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
		return false
	}
	return true
}
