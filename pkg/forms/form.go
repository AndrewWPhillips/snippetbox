package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// EmailRX can be used with the MatchesPattern (below) to check that a field looks like an email address
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Form is used in validation of HTML form fields to hold the field values and any validation errors
type Form struct {
	url.Values
	Errors errors
}

// New creates a Form used in validating a form.
// It's provided with a map with a key of the field name with a value of []string for zero or more field value(s)
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks that the specified field(s) are present and not an empty string
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MaxLength checks that a field does not exceed length requirements
func (f *Form) MaxLength(field string, d int) {
	if utf8.RuneCountInString(f.Get(field)) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// MinLength checks that a field is not too short but can still be empty (see Required above)
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

// PermittedValues checks that a field is found within a list of allowed values
// An empty string is allowed as a special case, assuming that Required (above) would be used if necessary
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" { // TODO: decide iof this if is really nec.
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

// MatchesPattern checks that a field matches a regex.
// It allows empty fields - to check for that use Required (above)
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// Valid returns true if there were no errors in validating the form
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
