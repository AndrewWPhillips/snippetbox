package forms

// Define a new errors type, which we will use to hold the validation error
// messages for forms. The name of the form field will be used as the key in
// this map.

// errors holds the error(s) found in validating the fields of a form
// The map key is the field name and the string slice holds the text of the error(s)
type errors map[string][]string

// Add adds an error message for the specified field
// It may be called multiple times to add more than one error to a specific field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves the first error for a field or an empty string if there are no errors
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
