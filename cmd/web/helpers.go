package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/models"
	"github.com/justinas/nosurf"
)

// serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound is a convenience wrapper around clientError which sends a 404 Not Found response
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// addDefaultData sets templateData field(s) that are used in rendering all (or several different) pages
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CSRFToken = nosurf.Token(r)                  // used on all pages that have forms
	td.AuthenticatedUser = app.authenticatedUser(r) // shown next to Logout button
	td.CurrentYear = time.Now().Year()              // shown on all pages as an example of dynamic data
	td.Flash = app.session.PopString(r, "flash")    // all pages can display (once only) a "flash" message
	return td
}

// render executes a HTML template for an endpoint handler
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// REFACTOR: The following code was replaced to better handle errors (see PROBLEM below)
	//// Execute the template set, passing in any dynamic data.
	//err := ts.Execute(w, td)
	//if err != nil {
	//	//  PROBLEM: this has problems as the response may have already been partially written (200 OK)
	//	//  and this also indirectly calls http.Error which calls WriteHeader (logs a message:
	//	// "http: superfluous response.WriteHeader call").
	//	app.serverError(w, err)
	//}

	// Write the response to a buffer, so we don't write a partial response and can send proper HTTP error
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Template executed OK so send the result (200 OK)
	if _, err = buf.WriteTo(w); err != nil {
		app.serverError(w, err)
		return
	}
}

// authenticatedUser returns info on the current user or nil if nobody is logged in
func (app *application) authenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}
