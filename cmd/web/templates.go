package main

import (
	"html/template"
	"log"
	"path/filepath"
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/forms"
	"github.com/andrewwphillips/snippetbox/pkg/models"
)

// templateData holds refs to any data that we want to pass to templates
type templateData struct {
	//AuthenticatedUser int          // ID
	AuthenticatedUser *models.User // user info or nil if not logged in
	CSRFToken         string
	CurrentYear       int
	Flash             string // used to display a "flash" message
	Form              *forms.Form
	Snippet           *models.Snippet
	Snippets          []*models.Snippet
}

// humanDate returns a nicely formatted string representation (UTC) of a time.Time object
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// functions is a map  of functions (for use in HTML templates) indexed by a name (string)
// Note that each function can only return one value (and optional error)
var functions = template.FuncMap{
	"humanDate": humanDate,
}

const (
	// pageTemplates etc contain the file name patterns for all the template
	// (.tmpl) files used to generate the web pages
	pageTemplates    = "*.page.tmpl"    // Defines the pages (home, login, etc)
	layoutTemplates  = "*.layout.tmpl"  // Base template(s) used for structure of all pages (just "base" for now)
	partialTemplates = "*.partial.tmpl" // Parts of pages used by more than one page (just footer for now)
)

// newTemplateCache loads and processes all the templates into a map for faster rendering
// of web pages.  The "dir" parameter is the path to the template files.  Files processed
// are those matched in the above const strings
func newTemplateCache(dir string) map[string]*template.Template {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Get the names of all the page template files
	pages, err := filepath.Glob(filepath.Join(dir, pageTemplates))
	if err != nil {
		log.Fatal(err)
	}
	if pages == nil {
		log.Fatal("No template files found at:", dir)
	}

	// Process each page template and save a ptr to the parsed template.Template
	for _, page := range pages {
		name := filepath.Base(page) // get page identifier

		// Parse the template, adding any custom functions
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			log.Fatal(err)
		}

		// We will need a layout (eg base)
		ts, err = ts.ParseGlob(filepath.Join(dir, layoutTemplates))
		if err != nil {
			log.Fatal(err)
		}

		// Add any partials that may be required
		ts, err = ts.ParseGlob(filepath.Join(dir, partialTemplates))
		if err != nil {
			log.Fatal(err)
		}

		cache[name] = ts // save in the returned map
	}

	return cache
}
