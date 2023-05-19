package forms

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/rs/zerolog/log"
	"github.com/xremming/abborre/esox/csrf"
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

	if name == "" {
		panic("Field name cannot be empty.")
	}

	if name == "_csrf" {
		panic("Field name cannot be _csrf.")
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

func (f FormBuilder) Empty(ctx context.Context) Form {
	if !f.done {
		panic("FormBuilder must be marked as done before parsing.")
	}

	out := Form{f.fieldOrdering, make(map[string]Field, len(f.fields)), nil}

	csrfStruct := csrf.FromContext(ctx)
	if csrfStruct != nil {
		out.fieldOrdering = append(out.fieldOrdering, "_csrf")
		out.fields["_csrf"] = Field{
			Name:  "_csrf",
			Kind:  KindHidden,
			Value: csrfStruct.Generate(),
		}
	}

	for name, field := range f.fields {
		out.fields[name] = field
	}

	return out
}

func lengthErrors(minLength, maxLength int, required bool, value string) []string {
	var out []string

	if len(value) == 0 && !required {
		return out
	}

	if minLength > 0 && len(value) < minLength {
		out = append(out, fmt.Sprintf("Must be at least %d characters.", minLength))
	}

	if maxLength > 0 && len(value) > maxLength {
		out = append(out, fmt.Sprintf("Must be at most %d characters.", maxLength))
	}

	return out
}

func (f FormBuilder) Parse(ctx context.Context, form url.Values) (Form, map[string]any) {
	if !f.done {
		panic("FormBuilder must be marked as done before parsing.")
	}

	logger := log.Ctx(ctx).With().Interface("form", form).Logger()

	out := Form{f.fieldOrdering, make(map[string]Field, len(f.fields)), nil}
	data := make(map[string]any)

	if form == nil {
		return out, data
	}

	csrfStruct := csrf.FromContext(ctx)
	if csrfStruct != nil {
		err := csrfStruct.Validate(ctx, form.Get("_csrf"))
		if err != nil {
			logger.Err(err).Msg("CSRF token validation failed.")

			if errors.Is(err, csrf.ErrTokenExpired) {
				out.Errors = append(out.Errors, "Form has expired, please retry.")
			} else {
				out.Errors = append(out.Errors, "Could not validate form, please retry.")
			}
		}
	}

	for name, field := range f.fields {
		value := form.Get(name)
		if field.shouldTrim() {
			value = strings.TrimSpace(value)
		}

		if field.Required && value == "" {
			field.Errors = append(field.Errors, "This field is required.")
		}

		field.Value = value

		switch field.Kind {
		case KindText:
			c := field.Config.(TextConfig)

			field.Errors = append(field.Errors, lengthErrors(c.MinLength, c.MaxLength, field.Required, value)...)

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				field.Errors = append(field.Errors, err...)
				data[name] = v
			}

		case KindPassword:
			c := field.Config.(PasswordConfig)

			field.Errors = append(field.Errors, lengthErrors(c.MinLength, c.MaxLength, field.Required, value)...)

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				field.Errors = append(field.Errors, err...)
				data[name] = v
			}

		case KindHidden:
			c := field.Config.(HiddenConfig)
			if c.Value != "" {
				field.Value = c.Value
			}

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				field.Errors = append(field.Errors, err...)
				data[name] = v
			}

		case KindDateTimeLocal:
			c := field.Config.(DateTimeLocalConfig)
			var (
				v   time.Time
				err error
			)

			if c.Location == "" {
				v, err = time.Parse(datetimeLocalFormat, value)
			} else {
				location, errLocation := time.LoadLocation(c.Location)
				if errLocation != nil {
					logger.Err(errLocation).Str("location", c.Location).Msg("invalid location")
					field.Errors = append(field.Errors, "Invalid location.")
				} else {
					v, err = time.ParseInLocation(datetimeLocalFormat, value, location)
				}
			}

			if err != nil {
				field.Errors = append(field.Errors, "Invalid date/time format.")
			} else {
				if !c.Min.IsZero() && v.Before(c.Min) {
					field.Errors = append(field.Errors, "Date/time must not be before "+c.Min.Format(datetimeLocalFormat)+".")
				}

				if !c.Max.IsZero() && v.After(c.Max) {
					field.Errors = append(field.Errors, "Date/time must not be after "+c.Max.Format(datetimeLocalFormat)+".")
				}
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
				field.Errors = append(field.Errors, fmt.Sprintf("%s is not a valid selection.", value))
			}

			if c.Parse == nil {
				data[name] = value
			} else {
				v, err := c.Parse(value)
				field.Errors = append(field.Errors, err...)
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
					field.Errors = append(field.Errors, fmt.Sprintf("%s is not a valid selection.", value))
				}

				if c.Parse == nil {
					dataValues = append(dataValues, value)
				} else {
					v, err := c.Parse(value)
					field.Errors = append(field.Errors, err...)
					dataValues = append(dataValues, v)
				}
			}

			data[name] = dataValues
		}

		out.fields[name] = field
	}

	return out, data
}
