package forms

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type FormBuilder struct {
	fieldOrdering []string
	fields        map[string]Field
	done          bool
}

func New() FormBuilder {
	return FormBuilder{
		fields: make(map[string]Field),
	}
}

type FieldConfigBuilder interface {
	Build(name string) Field
}

func (f FormBuilder) Field(name string, fieldBuilder FieldConfigBuilder) FormBuilder {
	if f.done {
		panic("FormBuilder cannot be modified after being marked as done.")
	}

	field := fieldBuilder.Build(name)
	f.fieldOrdering = append(f.fieldOrdering, field.Name)
	f.fields[field.Name] = field
	return f
}

func (f FormBuilder) Done() FormBuilder {
	f.done = true
	return f
}

func (f FormBuilder) Empty() Form {
	if !f.done {
		panic("FormBuilder must be marked as done before parsing.")
	}

	return Form{f.fieldOrdering, f.fields, nil}
}

func (f FormBuilder) Parse(form url.Values) (Form, map[string]any) {
	if !f.done {
		panic("FormBuilder must be marked as done before parsing.")
	}

	out := Form{f.fieldOrdering, make(map[string]Field, len(f.fields)), nil}
	for name, field := range f.fields {
		out.fields[name] = field
	}

	data := make(map[string]any)

	if form == nil {
		return out, data
	}

	for name, field := range f.fields {
		value := form.Get(name)
		if field.shouldTrim() {
			value = strings.TrimSpace(value)
		}

		if field.Required && value == "" {
			out.addFieldErrors(name, "This field is required.")
		}

		out.setFieldValue(name, value)

		switch field.Kind {
		case KindText:
			c := field.Config.(TextConfig)

			if c.MinLength > 0 && len(value) < c.MinLength {
				out.addFieldErrors(name, fmt.Sprintf("Must be at least %d characters.", c.MinLength))
			}

			if c.MaxLength > 0 && len(value) > c.MaxLength {
				out.addFieldErrors(name, fmt.Sprintf("Must be at most %d characters.", c.MaxLength))
			}

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				out.addFieldErrors(name, err...)
				data[name] = v
			}

		case KindPassword:
			c := field.Config.(PasswordConfig)

			if c.MinLength > 0 && len(value) < c.MinLength {
				out.addFieldErrors(name, fmt.Sprintf("Must be at least %d characters.", c.MinLength))
			}

			if c.MaxLength > 0 && len(value) > c.MaxLength {
				out.addFieldErrors(name, fmt.Sprintf("Must be at most %d characters.", c.MaxLength))
			}

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				out.addFieldErrors(name, err...)
				data[name] = v
			}

		case KindHidden:
			c := field.Config.(HiddenConfig)
			out.setFieldValue(name, c.Value)

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				out.addFieldErrors(name, err...)
				data[name] = v
			}

		case KindDateTimeLocal:
			c := field.Config.(DateTimeLocalConfig)
			var (
				v   time.Time
				err error
			)

			if c.Location == nil {
				v, err = time.Parse(datetimeLocalFormat, value)
			} else {
				v, err = time.ParseInLocation(datetimeLocalFormat, value, c.Location)
			}

			if err != nil {
				out.addFieldErrors(name, "Invalid date/time format.")
			}

			if !c.Min.IsZero() && v.Before(c.Min) {
				out.addFieldErrors(name, "Date/time must not be before "+c.Min.Format(datetimeLocalFormat)+".")
			}

			if !c.Max.IsZero() && v.After(c.Max) {
				out.addFieldErrors(name, "Date/time must not be after "+c.Max.Format(datetimeLocalFormat)+".")
			}

			data[name] = v

		case KindSelect:
			c := field.Config.(SelectConfig)
			found := false
			for _, option := range c.Options {
				if option.Value == value {
					found = true
					break
				}
			}

			if !found {
				out.addFieldErrors(name, fmt.Sprintf("%s is not a valid selection.", value))
			}

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				out.addFieldErrors(name, err...)
				data[name] = v
			}

		case KindSelectMulti:
			values := form[name]
			c := field.Config.(SelectMultiConfig)
			dataValues := make([]any, 0, len(values))

			for _, value := range values {
				found := false
				for _, option := range c.Options {
					if option.Value == value {
						found = true
						break
					}
				}

				if !found {
					out.addFieldErrors(name, fmt.Sprintf("%s is not a valid selection.", value))
				}

				if c.Parse == nil {
					dataValues = append(dataValues, value)
				} else {
					v, err := c.Parse(value)
					out.addFieldErrors(name, err...)
					dataValues = append(dataValues, v)
				}
			}

			data[name] = dataValues
		}
	}

	return out, data
}
