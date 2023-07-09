package render

import (
	"gofaster/internal/bytesconv"
	"html/template"
	"net/http"
)

type HTMLRender struct {
	Template *template.Template
}

type HTML struct {
	Data       any
	Name       string
	Template   *template.Template
	IsTemplate bool
}

func (H *HTML) Render(w http.ResponseWriter) error {
	H.WriteContentType(w)
	if H.IsTemplate {
		err := H.Template.ExecuteTemplate(w, H.Name, H.Data)
		return err
	}
	_, err := w.Write(bytesconv.StringToBytes(H.Data.(string)))
	return err
}

func (H *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
