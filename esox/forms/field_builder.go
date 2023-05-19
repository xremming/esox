package forms

import (
	"fmt"
	"strconv"
	"time"
)

func ParseDuration(value string) (any, []string) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return nil, []string{err.Error()}
	}

	return duration, nil
}

type ParseFunc func(string) (any, []string)

type TextConfig struct {
	Parse     ParseFunc
	Multiline bool

	Pattern   string
	MinLength int
	MaxLength int
}

type PasswordConfig struct {
	Parse     ParseFunc
	Pattern   string
	MinLength int
	MaxLength int
}

type HiddenConfig struct {
	Parse ParseFunc
	Value string
}

type DateTimeLocalConfig struct {
	Min      time.Time
	Max      time.Time
	Location *time.Location
}

type OptionConfig struct {
	Value    string
	Label    string
	Selected bool
}

type SelectConfig struct {
	Parse   ParseFunc
	Radio   bool
	Options []OptionConfig
}

type SelectMultiConfig struct {
	Parse    ParseFunc
	Checkbox bool
	Options  []OptionConfig
}

type FieldConfig interface {
	TextConfig | PasswordConfig | DateTimeLocalConfig | SelectConfig | SelectMultiConfig
}

type FieldBuilder[C FieldConfig] struct {
	Label    string
	Required bool
	NoTrim   bool
	Config   C
}

func (fb FieldBuilder[C]) Build(name string) Field {
	id := idFromName(name)

	var (
		fieldType FieldKind
		options   []Option
	)

	switch params := any(fb.Config).(type) {
	case TextConfig:
		fieldType = KindText

	case PasswordConfig:
		fieldType = KindPassword

	case DateTimeLocalConfig:
		fieldType = KindDateTimeLocal

	case SelectConfig:
		fieldType = KindSelect
		for i, option := range params.Options {
			options = append(options, Option{
				ID:       idFromName(name, strconv.Itoa(i)),
				Value:    option.Value,
				Label:    option.Label,
				Selected: option.Selected,
			})
		}

	case SelectMultiConfig:
		fieldType = KindSelectMulti
		for i, option := range params.Options {
			options = append(options, Option{
				ID:       idFromName(name, strconv.Itoa(i)),
				Value:    option.Value,
				Label:    option.Label,
				Selected: option.Selected,
			})
		}

	default:
		panic(fmt.Sprintf("FieldBuilder[P].Build is missing a case for type %T", fb.Config))
	}

	return Field{
		ID:    id,
		Name:  name,
		Label: fb.Label,

		Required: fb.Required,
		NoTrim:   fb.NoTrim,

		Kind:    fieldType,
		Options: options,
		Config:  fb.Config,
	}
}
