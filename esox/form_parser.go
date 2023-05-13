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

	// Layout to use for parsing the time. Defaults to "2006-01-02T15:04".
	Layout string
}

func (f *FormParser) ParseTime(form url.Values, fieldName string, opts ParseTimeOpts) time.Time {
	value := form.Get(fieldName)
	value = strings.TrimSpace(value)
	if opts.Required && value == "" {
		f.AddFieldError(fieldName, "Must not be empty.")
		return time.Time{}
	} else if value == "" {
		return time.Time{}
	}

	layout := opts.Layout
	if layout == "" {
		layout = "2006-01-02T15:04"
	}

	var (
		valueParsed time.Time
		err         error
	)

	if opts.Location != nil {
		valueParsed, err = time.ParseInLocation(layout, value, opts.Location)
	} else {
		valueParsed, err = time.Parse(layout, value)
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

type ParseDurationOpts struct {
	Required bool
}

func (f *FormParser) ParseDuration(form url.Values, fieldName string, opts ParseDurationOpts) time.Duration {
	value := form.Get(fieldName)
	value = strings.TrimSpace(value)
	if opts.Required && value == "" {
		f.AddFieldError(fieldName, "Must not be empty.")
		return 0
	}

	valueParsed, err := time.ParseDuration(value)
	if err != nil {
		f.AddFieldError(fieldName, "Invalid duration format.")
		return 0
	}

	return valueParsed
}

func (f *FormParser) ParseDurationPointer(form url.Values, fieldName string, opts ParseDurationOpts) *time.Duration {
	res := f.ParseDuration(form, fieldName, opts)
	if res == 0 {
		return nil
	}

	return &res
}
