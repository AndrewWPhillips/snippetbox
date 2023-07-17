package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// routes creates the router (mux) to be used with the HTTPS server and adds all
// the different endpoint and middleware handlers.  The parameter (root) is the
// root disk location of all static files that need to be served (CSS, JS, PNG etc).
func (app *application) routes(root string) http.Handler {
	// standardMiddleware is used on all routes
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// dynamicMiddleware is used for anything that needs session info and/or uses forms
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	// REFACTOR: Replaced std lib router with pat for extra features such as
	//  * named capture - ":id" in the "/snippet/:id" route
	//  * separate GET and POST handlers for "/snippet/create"
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", app.home)
	//mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	//mux.HandleFunc("/snippet", app.showSnippet)
	//mux.HandleFunc("/snippet/create", app.createSnippet)

	// REFACTOR2: Changed the next 6 lines to use dynamicMiddleware (for session management)
	//mux := pat.New()
	//mux.Get("/", http.HandlerFunc(app.home))
	//mux.Get("/snippet/create", http.HandlerFunc(app.createSnippetForm))
	//mux.Post("/snippet/create", http.HandlerFunc(app.createSnippet))
	//mux.Get("/snippet/:id", http.HandlerFunc(app.showSnippet)) // must be after "/snippet/create"
	//mux.Get("/static/", http.StripPrefix("/static", fileServer))

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet)) // must be after "/snippet/create" in this list
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	if root != "" {
		// Serve files used in the UI from /static/ path using std lib file server.
		// Note that the path given is relative to the project directory root.
		fileServer := http.FileServer(http.Dir(root))
		mux.Get("/static/", http.StripPrefix("/static", fileServer))
	}
	mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}
