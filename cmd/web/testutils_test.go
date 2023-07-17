package main

import (
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/models/mock"
	"github.com/golangcollege/sessions"
)

// newTestApplication creates an app instance with mock dependencies
func newTestApplication(t *testing.T) *application {
	// Create a session manager instance, with the same settings as production.
	session := sessions.New([]byte("3dSm5MnygFHh7XidAtbskXrjbwfoJcbJ"))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	// Initialize the dependencies, using the mocks for the loggers and
	// database models.
	return &application{
		errorLog:      log.New(io.Discard, "", 0),
		infoLog:       log.New(io.Discard, "", 0),
		session:       session,
		snippets:      mock.NewSnippetModel(""),
		templateCache: newTemplateCache("./../../ui/html/"),
		users:         mock.NewUserModel(""),
	}
}

// testServer embeds a httptest.Server to add useful methods
type testServer struct {
	*httptest.Server
}

// newTestServer returns a new instance of our custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	server := httptest.NewTLSServer(h)

	// Allow cookies to be saved & automatically sent in subsequent requests.
	// (This will be (at least) required for tests involving sessions.)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	server.Client().Jar = jar

	// Disable redirects, so we get the 3XX response & not the redirected response
	server.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{server}
}

// get makes a GET request to the test server & returns status, headers and body
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	// This simulates such things as the user clicking on a link in a browser to display a page
	response, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err2 := io.ReadAll(response.Body)
	if err2 != nil {
		t.Fatal(err2)
	}

	return response.StatusCode, response.Header, string(body)
}

// postForm makes a POST request as though coming from a form submissions
// The 3rd parameter (of type url.Values) contains the values of the form fields
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, []byte) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body.
	defer rs.Body.Close()
	body, err2 := io.ReadAll(rs.Body)
	if err2 != nil {
		t.Fatal(err2)
	}

	// Return the response status, headers and body.
	return rs.StatusCode, rs.Header, body
}

// regex to extract CSRF token from an HTML form
var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

// extractCSRFToken gets the CSRF token from the hidden input field found in all our HTML forms
func extractCSRFToken(t *testing.T, body []byte) string {
	// This finds the line with the hidden CSRF field (in a form in the HTML text) and returns a
	// slice with 2 strings (each string is stored as []byte) where
	//   matches[0] contains the full matching line containing the form field
	//   matches[1] contains the part of the line that matches the capture group (.+)
	// Since there is only one capture group - one set of parentheses - in the above regex then len(matches) == 2
	matches := csrfTokenRX.FindSubmatch(body)
	if matches == nil {
		t.Fatal("no csrf token found in body")
	}
	if len(matches) != 2 {
		t.Fatal("internal error finding csrf token")
	}

	// Since the rendered HTML may have escaped special characters (so it
	// displays properly in browser) we need to restore or "unescape" them in
	// the returned (base64 encoded) CSRF token just in case it has any of them
	return html.UnescapeString(string(matches[1]))
}
