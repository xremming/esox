package forms

type FieldKind string

const (
	KindText          FieldKind = "text"
	KindPassword      FieldKind = "password"
	KindHidden        FieldKind = "hidden"
	KindDateTimeLocal FieldKind = "datetime-local"
	KindSelect        FieldKind = "select"
	KindSelectMulti   FieldKind = "select-multi"
)

type Option struct {
	ID       string
	Value    string
	Label    string
	Selected bool
}

type Field struct {
	ID    string
	Name  string
	Label string

	Value  string
	Errors []string

	Required bool
	NoTrim   bool

	Kind    FieldKind
	Options []Option
	Config  any
}

func (f Field) shouldTrim() bool {
	if f.NoTrim {
		return false
	}

	switch f.Kind {
	case KindPassword, KindHidden, KindSelect, KindSelectMulti:
		return false
	}

	return true
}
