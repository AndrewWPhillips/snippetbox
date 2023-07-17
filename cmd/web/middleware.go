package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

// contextKey is our unique type to ensure our context.Context key for user
// data is unique - if we just used a string it could clash with middleware
// from imported packages that also uses a context key of "user"
type contextKey string

// contextKeyUser is the key used with response context to obtain the current user
const (
	sessionUserID  = "userID"           // key for user ID stored in the session (cookie?)
	contextKeyUser = contextKey("user") // key for user details stored in the context.Context
)

// authenticate adds middleware that checks for the session "userID" and (if found)
// looks the user in the "Users" table and adds it to the request context using the
// custom context key contextKeyUser.  If the session has a "userID" but there is no
// DB record for the user then the session "userID" is removed
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check is user is logged in
		exists := app.session.Exists(r, sessionUserID)
		if !exists {
			next.ServeHTTP(w, r) // not logged in so continue to next in chain (no auth)
			return
		}

		user, err := app.users.Get(app.session.GetInt(r, sessionUserID))
		if err != nil {
			app.serverError(w, err)
			return
		} else if user == nil { // err == nil and user == nil means not found
			// user must have been removed so log them out and continue (no auth)
			app.session.Remove(r, sessionUserID)
			next.ServeHTTP(w, r)
			return
		}

		// Add user data to the request's context
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAuthenticatedUser blocks requests unless the user is logged in
func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If not logged in redirect to the login page (and don't call next.ServeHTTP)
		if app.authenticatedUser(r) == nil {
			// REFACTOR:I replaced the following line (which just sends the user
			// back to the login page) with an auth error response
			//http.Redirect(w, r, "/user/login", http.StatusFound)
			http.Error(w, "Not logged in", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// logRequest logs all requests to stdout
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

// recoverPanic sends a 500 (internal server error) if there is a panic in the next handler(s)
// (This avoids the poor default behaviour which just terminates the connection.)
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				// We only get here if the handler (see ServeHTTP below) panics
				w.Header().Set("Connection", "close") // trigger connection close
				app.serverError(w, fmt.Errorf("Error: %v", panicValue))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// secureHeaders adds HTTP headers to help prevent XSS and Clickjacking attacks
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

// noSurf is middleware used to protect forms from CSRF attacks
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}
