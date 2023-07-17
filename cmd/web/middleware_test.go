package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSecureHeaders checks that the middleware returned by secureHeaders correctly
// sets the headers and calls the next handler in the chain.
func TestSecureHeaders(t *testing.T) {
	const expectedResponse = "OK (next response)" // body returned by "next" handler

	// Initialize a new httptest.ResponseRecorder and dummy http.Request.
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get dummy handler (returns OK & expectedResponse) used as middleware's "next" handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expectedResponse))
	})

	// Call the middleware handler function and get the recorded result
	secureHeaders(next).ServeHTTP(recorder, req)
	result := recorder.Result()

	// Check it has correctly set X-Frame-Options and X-XSS-Protection headers
	if hdr := result.Header.Get("X-Frame-Options"); hdr != "deny" {
		t.Errorf("want %q; got %q", "deny", hdr)
	}
	if hdr := result.Header.Get("X-XSS-Protection"); hdr != "1; mode=block" {
		t.Errorf("want %q; got %q", "1; mode=block", hdr)
	}

	// Check that the next handler was called correctly
	if result.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, result.StatusCode)
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	if body := string(body); body != expectedResponse {
		t.Errorf("expected next body to equal %q but got %.40q (len %d)", expectedResponse, body, len(body))
	}
}
