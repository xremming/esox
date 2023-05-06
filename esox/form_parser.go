package esox

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type FormParser struct {
	Errors      []string
	FieldErrors map[string][]string
}

func NewFormParser() *FormParser {
	return &FormParser{
		Errors:      make([]string, 0),
		FieldErrors: make(map[string][]string),
	}
}

func (f *FormParser) AddError(err string) {
	f.Errors = append(f.Errors, err)
}

func (f *FormParser) AddFieldError(fieldName, err string) {
	f.FieldErrors[fieldName] = append(f.FieldErrors[fieldName], err)
}

func (f *FormParser) HasErrors() bool {
	if len(f.Errors) > 0 {
		return true
	}

	for _, fieldErrors := range f.FieldErrors {
		if len(fieldErrors) > 0 {
			return true
		}
	}

	return false
}

func (f *FormParser) UpdateForm(form *FormData) {
	form.Errors = append(form.Errors, f.Errors...)
	for _, field := range form.Fields {
		fieldErrors, ok := f.FieldErrors[field.Name]
		if !ok {
			continue
		}
		field.Errors = append(field.Errors, fieldErrors...)
	}
}

type ParseStringOpts struct {
	Required bool
	NoTrim   bool
	Matches  *regexp.Regexp
}

func (f *FormParser) ParseString(form url.Values, fieldName string, opts ParseStringOpts) string {
	value := form.Get(fieldName)
	if !opts.NoTrim {
		value = strings.TrimSpace(value)
	}

	if opts.Required && value == "" {
		f.AddFieldError(fieldName, "Must not be empty.")
	}

	if opts.Matches != nil && !opts.Matches.MatchString(value) {
		f.AddFieldError(fieldName, fmt.Sprintf("Must match the regex %s.", opts.Matches))
	}

	return value
}

func (f *FormParser) ParseStringPointer(form url.Values, fieldName string, opts ParseStringOpts) *string {
	value := f.ParseString(form, fieldName, opts)
	if value == "" {
		return nil
	}

	return &value
}

type ParseTimeOpts struct {
	Required bool
	Location *time.Location
}

// ParseTime parses a datetime in format '2006-01-02T15:04'.
func (f *FormParser) ParseTime(form url.Values, fieldName string, opts ParseTimeOpts) time.Time {
	value := form.Get(fieldName)
	value = strings.TrimSpace(value)
	if opts.Required && value == "" {
		f.AddFieldError(fieldName, "Must not be empty.")
		return time.Time{}
	} else if value == "" {
		return time.Time{}
	}

	var (
		valueParsed time.Time
		err         error
	)

	if opts.Location != nil {
		valueParsed, err = time.ParseInLocation("2006-01-02T15:04", value, opts.Location)
	} else {
		valueParsed, err = time.Parse("2006-01-02T15:04", value)
	}

	if err != nil {
		f.AddFieldError(fieldName, "Invalid date format.")
		return time.Time{}
	}

	return valueParsed
}

func (f *FormParser) ParseTimePointer(form url.Values, fieldName string, opts ParseTimeOpts) *time.Time {
	res := f.ParseTime(form, fieldName, opts)
	if res.IsZero() {
		return nil
	}

	return &res
}
