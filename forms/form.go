package forms

type Form struct {
	fieldOrdering []string
	fields        map[string]Field

	Errors []string
}

func (f *Form) Field(name string) *Field {
	field, ok := f.fields[name]
	if !ok {
		return nil
	}

	return &field
}

func (f *Form) Fields() []Field {
	fields := make([]Field, 0, len(f.fieldOrdering))
	for _, name := range f.fieldOrdering {
		fields = append(fields, f.fields[name])
	}

	return fields
}

func (f *Form) HasErrors() bool {
	if len(f.Errors) > 0 {
		return true
	}

	for _, field := range f.Fields() {
		if len(field.Errors) > 0 {
			return true
		}
	}

	return false
}
