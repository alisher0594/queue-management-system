package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// EmailRX is a regex for email validation
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// PhoneRX is a regex for phone number validation
var PhoneRX = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// Form represents a form with validation
type Form struct {
	url.Values
	Errors errors
}

// errors holds validation error messages
type errors map[string][]string

// Add adds an error message for a given field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get gets the first error message for a given field
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}

// New creates a new form
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks that specific fields are not blank
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MaxLength checks that a field is not longer than the specified length
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// MinLength checks that a field is at least the specified length
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

// MatchesPattern checks that a field matches a regular expression
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// PermittedValues checks that a field value is in a list of permitted values
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

// Valid returns true if there are no validation errors
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Get gets a form value by key, with a fallback
func (f *Form) Get(key string) string {
	vs := f.Values[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// GetInt gets a form value as an integer
func (f *Form) GetInt(key string) (int, error) {
	str := f.Get(key)
	if str == "" {
		return 0, fmt.Errorf("field %s is required", key)
	}
	return strconv.Atoi(str)
}
