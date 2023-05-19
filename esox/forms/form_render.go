package forms

import (
	"html/template"

	"github.com/xremming/abborre/esox/utils"
)

var divTemplate = `
{{ range .Form.Fields }}
<div class="{{ $.FieldClass }}">

  <label for="{{ .ID }}">{{ .Label }}{{ if .Required }}*{{ end }}</label>

  {{ if eq .Kind "text" }}
    {{ if .Config.Multiline }}
    <textarea id="{{ .ID }}" name="{{ .Name }}">{{ .Value }}</textarea>
    {{ else }}
    <input id="{{ .ID }}" type="text" name="{{ .Name }}" value="{{ .Value }}" />
    {{ end }}
  {{ else if eq .Kind "datetime-local" }}
  <input id="{{ .ID }}" type="datetime-local" name="{{ .Name }}" value="{{ .Value }}" />
  {{ else if eq .Kind "select" }}
  <select id="{{ .ID }}" name="{{ .Name }}">
  {{ range .Options }}
    <option value="{{ .Value }}"{{ if .Selected }} selected{{ end }}>{{ .Label }}</option>
  {{ end }}
  </select>
  {{ end }}

  {{ range .Errors }}
  <div class="{{ $.ErrorClass }}">{{ . }}</div>
  {{ end }}

  </div>
{{ end }}

{{ range .Form.Errors }}
<div class="{{ $.ErrorClass }}">{{ . }}</div>
{{ end }}
`

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
