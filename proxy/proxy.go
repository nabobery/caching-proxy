package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// StartServer starts the caching proxy server forwarding requests from the given port to the origin.
func StartServer(origin string, port int) {
	originURL, err := url.Parse(origin)
	if err != nil {
		log.Fatalf("Error parsing origin URL: %v", err)
	}

	// Director rewrites the request to forward it to the origin.
	director := func(req *http.Request) {
		req.URL.Scheme = originURL.Scheme
		req.URL.Host = originURL.Host
		// Log the new URL.
		log.Printf("[Director] Forwarding request to: %s", req.URL.String())

		// Remove proxy-specific headers.
		req.Header.Del("X-Forwarded-For")
		req.Header.Del("X-Forwarded-Host")
		req.Header.Del("X-Forwarded-Proto")
		req.Header.Del("Via")
		req.Header.Del("Forwarded")

		// Use the origin's host.
		req.Host = originURL.Host
	}

	// Create a ReverseProxy with our custom transport and a ModifyResponse callback.
	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: &CachingTransport{Transport: http.DefaultTransport},
		ModifyResponse: func(resp *http.Response) error {
			log.Printf("[ModifyResponse] Response received with status: %d", resp.StatusCode)
			// Log headers before modification.
			for key, values := range resp.Header {
				log.Printf("[ModifyResponse] Before - Header %s: %v", key, values)
			}

			// Remove hop-by-hop headers.
			hopByHopHeaders := []string{"Connection", "Keep-Alive", "Proxy-Authenticate",
				"Proxy-Authorization", "TE", "Trailer", "Transfer-Encoding", "Upgrade"}
			for _, h := range hopByHopHeaders {
				if v := resp.Header.Get(h); v != "" {
					log.Printf("[ModifyResponse] Removing hop-by-hop header: %s", h)
					resp.Header.Del(h)
				}
			}

			// Force the X-Cache header.
			xcache := resp.Header.Get("X-Cache")
			if xcache == "" {
				log.Printf("[ModifyResponse] X-Cache header missing — forcing it to MISS")
				xcache = "MISS"
			} else {
				log.Printf("[ModifyResponse] Found X-Cache header: %s", xcache)
			}
			resp.Header.Set("X-Cache", xcache)

			// Log final headers.
			log.Printf("[ModifyResponse] Final response headers:")
			for k, v := range resp.Header {
				log.Printf("[ModifyResponse] %s: %v", k, v)
			}

			return nil
		},
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting caching proxy on %s forwarding to %s", addr, origin)
	if err := http.ListenAndServe(addr, proxy); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
