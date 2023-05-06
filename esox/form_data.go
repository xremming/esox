package esox

type FormField struct {
	Name   string
	Errors []string
}

type FormData struct {
	Errors []string
	Fields []FormField
}

func (f FormData) Field(name string) *FormField {
	for _, field := range f.Fields {
		if field.Name == name {
			return &field
		}
	}

	return nil
}
