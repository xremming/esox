package forms

import (
	_ "embed"
	"html/template"

	"github.com/xremming/abborre/esox/utils"
)

//go:embed templates/form_div.html
var divTemplate string

var div = template.Must(template.New("div").Parse(divTemplate))

func (f *Form) RenderDiv(fieldClass string, errorClass string) template.HTML {
	buf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(buf)

	err := div.Execute(buf, struct {
		FieldClass string
		ErrorClass string
		Form       *Form
	}{
		FieldClass: fieldClass,
		ErrorClass: errorClass,
		Form:       f,
	})
	if err != nil {
		panic(err)
	}

	return template.HTML(buf.String())
}
