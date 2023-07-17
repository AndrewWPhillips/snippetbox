package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/andrewwphillips/snippetbox/pkg/forms"
	"github.com/andrewwphillips/snippetbox/pkg/models"
)

// home shows the default page - list of the latest snippets (see home.page.html)
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// This is not nec. when we moved from the std lib router to pat since (unlike other patterns
	// ending in a slash) "/" only matches itself
	//if r.URL.Path != "/" {
	//	app.notFound(w)
	//	return
	//}

	ss, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// NOTE: The following code was replaced by the render method for performance (all templates are now
	// "cached" at startup) and to avoid duplication of code that parses and executes the HTML template.

	//// Use the template.ParseFiles() function to read the template files into a template set.
	//// The home page (home.page.tmpl) must be the *first* file in the slice.
	//ts, err := template.ParseFiles([]string{
	//	"./ui/html/home.page.tmpl",
	//	"./ui/html/footer.partial.tmpl",
	//	"./ui/html/base.layout.tmpl",
	//}...)
	//if err != nil {
	//	app.serverError(w, err)
	//	return
	//}
	//
	//// Pass in the templateData struct when executing the template.
	//err = ts.Execute(w, &templateData{Snippets: ss})
	//if err != nil {
	//  // TODO: this has problems as the response may have already been partially written (200 OK)
	//  // and this also indirectly calls WriteHeader (http: superfluous response.WriteHeader)
	//	app.serverError(w, err)
	//}
	app.render(w, r, "home.page.tmpl", &templateData{Snippets: ss})
}

// showSnippet displays the "show" page to view a single snippet
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	//id, err := strconv.Atoi(r.URL.Query().Get("id")) // get "id" query param.
	id, err := strconv.Atoi(r.URL.Query().Get(":id")) // now using pat's named capture (part of the URL) not query parameter
	if err != nil || id < 1 {
		app.notFound(w) // not a number or -ve/zero
		return
	}

	s, err2 := app.snippets.Get(id)
	if err2 != nil {
		app.serverError(w, err)
		return
	}
	if s == nil {
		app.notFound(w) // not an actual snippet number
		return
	}

	// All this template stuff was factored out into app.render call (below)
	//files := []string{
	//	"./ui/html/show.page.tmpl",
	//	"./ui/html/base.layout.tmpl",
	//	"./ui/html/footer.partial.tmpl",
	//}
	//
	//ts, err := template.ParseFiles(files...)
	//if err != nil {
	//	app.serverError(w, err)
	//	return
	//}
	//
	//err = ts.Execute(w, templateData{Snippet: s})
	//if err != nil {
	//  // TODO: this has problems as the response may have already been partially written (200 OK)
	//  // and this also indirectly calls WriteHeader (http: superfluous response.WriteHeader)
	//	app.serverError(w, err)
	//}

	app.render(w, r, "show.page.tmpl", &templateData{Snippet: s})
}

// createSnippetForm displays a form to the user that allows them to create a new snippet
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	// We now need to send an empty Form so that the HTML form (between {{with .Form}} ... {{end}} is shown
	//app.render(w, r, "create.page.tmpl", nil)
	app.render(w, r, "create.page.tmpl", &templateData{Form: forms.New(nil)})
}

// createSnippet is a POST method that responds to the submission of the create snippet form
// It adds a snippet to the database and displays it (redirects to showSnippet)
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32768) // limit to 32K
	if err := r.ParseForm(); err != nil {
		fmt.Fprintln(w, err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	/* // This is the old "manual" validation - replaced
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")
	expires := r.PostForm.Get("expires")

	// Validate all the form fields, building a list of errors
	errors := make(map[string]string)
	if strings.TrimSpace(title) == "" {
		errors["title"] = "This field cannot be blank"
	} else if len := utf8.RuneCountInString(title); len > 100 {
		errors["title"] = fmt.Sprintf("This field is too long (%d characters) - the limit is 100", len)
	}

	// Check that the Content field isn't blank.
	if strings.TrimSpace(content) == "" {
		errors["content"] = "This field cannot be blank"
	}

	// Check the expires field isn't blank and matches one of the permitted
	// values ("1", "7" or "365").
	if strings.TrimSpace(expires) == "" {
		errors["expires"] = "This field cannot be blank"
	} else if expires != "365" && expires != "7" && expires != "1" {
		errors["expires"] = "This field is invalid"
	}

	// If there are any validation errors, re-display the page passing in the
	// validation errors and previously submitted r.PostForm data.
	if len(errors) > 0 {
		app.render(w, r, "create.page.tmpl", &templateData{
			FormErrors: errors,
			FormData:   r.PostForm,
		})
		return
	}

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	*/

	// Validate the form fields
	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")
	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	// Add a snippet using the (now validated) form fields
	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add a "flash" message to be displayed later
	app.session.Put(r, "flash", "Snippet successfully created!")

	// Redirect the user to a page showing the new snippet
	//http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther) // id value => ":id"
}

// signupUserForm displays a form to the user allowing them to create a login
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{Form: forms.New(nil)})
}

// signupUser is a POST method called in response to the signup Form, to create a new login
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintln(w, err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form fields
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	id, err2 := app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err2 == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err2 != nil {
		app.serverError(w, err2)
		return
	}

	_ = id // TODO use id to automatically login

	// Add a "flash" message to be displayed in the next page
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// loginUserForm displays a form allowing a user to login
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{Form: forms.New(nil)})
}

// loginUser is a POST method that responds to submission of the login form
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintln(w, err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form fields
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)
	if !form.Valid() {
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	}

	id, name, err2 := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err2 == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Invalid email or password")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err2 != nil {
		app.serverError(w, err2)
		return
	}

	// Login successful so add the ID to the session and display a message
	app.session.Put(r, sessionUserID, id)
	app.session.Put(r, "flash", "Hello "+name)
	http.Redirect(w, r, "/", http.StatusSeeOther) // home
}

// logoutUser is a POST method that logs out the user
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, sessionUserID)
	app.session.Put(r, "flash", "You have been logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

const pingResponse = "OK"

// ping is just used to check that the server is still responsive
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(pingResponse))
}
