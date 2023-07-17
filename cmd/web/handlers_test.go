package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestPing is a very simple test of the ping handler
func TestPing(t *testing.T) {
	// Create the test ResponseRecorder, and get the result of the ping handler
	recorder := httptest.NewRecorder()
	ping(recorder, nil)
	result := recorder.Result()

	// Check we got expected status and body
	if result.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, result.StatusCode)
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	if body := string(body); body != pingResponse {
		t.Errorf("expected body %q but got %.40q (len %d)", pingResponse, body, len(body))
	}
}

/*
// REFACTOR: replaced this (see below) after extracting code into reuseable test utils (see testutils_test.go)
func TestPingEndToEnd(t *testing.T) {
	app := &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
	}

	// Startup an HTTPS test server.
	server := httptest.NewTLSServer(app.routes(""))
	defer server.Close()

	// Send a ping to server (listening on a randomly-chosen port at server.URL)
	response, err := server.Client().Get(server.URL + "/ping")
	if err != nil {
		t.Fatal(err)
	}

	// Ensure we got expected status and body
	if response.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, response.StatusCode)
	}
	defer response.Body.Close()
	body, err2 := io.ReadAll(response.Body)
	if err2 != nil {
		t.Fatal(err2)
	}
	if body := string(body); body != pingResponse {
		t.Errorf("expected body %q but got %.40q (len %d)", pingResponse, body, len(body))
	}
}
*/

// TestPingEndToEnd tests that ping and routing work correctly
func TestPingEndToEnd(t *testing.T) {
	// Create a mock app and start test server
	app := newTestApplication(t)
	server := newTestServer(t, app.routes(""))
	defer server.Close()

	// Send the ping and check that we got expected status and body
	code, _, body := server.get(t, "/ping")
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}
	if body != pingResponse {
		t.Errorf("expected body %q but got %.40q (len %d)", pingResponse, body, len(body))
	}
}

// TestShowSnippet tests different requests for the HTML page to display a snippet
func TestShowSnippet(t *testing.T) {
	// Create a mock app and start test server
	app := newTestApplication(t)
	server := newTestServer(t, app.routes(""))
	defer server.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, "An old silent pond..."},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, http.StatusText(http.StatusNotFound)},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, http.StatusText(http.StatusNotFound)},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, http.StatusText(http.StatusNotFound)},
		{"String ID", "/snippet/foo", http.StatusNotFound, http.StatusText(http.StatusNotFound)},
		{"Empty ID", "/snippet/", http.StatusNotFound, "not found"},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates a GET request for the snippet page from the client
			// The body is the HTML that would be displayed (if code is 200)
			// If code is not 200 (4XX or 5XX) then body may be empty or contain an error message with more info
			code, _, body := server.get(t, tt.urlPath)

			// code can be OK (200) if the snippet was found else 404 (not found) - anything else is a bug
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("expected body to contain %q", tt.wantBody)
			}
		})
	}
}

// TestSignupUser tests requests for the signup page (form)
func TestSignupUser(t *testing.T) {
	app := newTestApplication(t)
	server := newTestServer(t, app.routes(""))
	defer server.Close()

	_, _, body := server.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, []byte(body))

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPa$$word", csrfToken, http.StatusSeeOther, nil},
		{
			"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field cannot be blank"),
		},
		{"Empty email", "Bob", "", "validPa$$word", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{
			"Empty password", "Bob", "bob@example.com", "", csrfToken, http.StatusOK,
			[]byte("This field cannot be blank"),
		},
		{
			"Invalid email (incomplete domain)", "Bob", "bob@example.", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field is invalid"),
		},
		{
			"Invalid email (missing @)", "Bob", "bobexample.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field is invalid"),
		},
		{
			"Invalid email (missing local part)", "Bob", "@example.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field is invalid"),
		},
		{
			"Short password", "Bob", "bob@example.com", "pa$$word", csrfToken, http.StatusOK,
			[]byte("This field is too short (minimum is 10 characters)"),
		},
		{
			"Duplicate email", "Bob", "alice@example.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("Address is already in use"),
		},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := server.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}
