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

func (f *Form) Set(name, value string) {
	f.setFieldValue(name, value)
}

func (f *Form) setFieldValue(name, value string) {
	field, ok := f.fields[name]
	if !ok {
		panic("cannot set value of non-existent field")
	}

	field.Value = value
	f.fields[name] = field
}

func (f *Form) addFieldErrors(name string, errors ...string) {
	field, ok := f.fields[name]
	if !ok {
		panic("cannot add errors to non-existent field")
	}

	field.Errors = append(field.Errors, errors...)
	f.fields[name] = field
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
