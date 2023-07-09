package render

import (
	"encoding/json"
	"net/http"
)

type JSON struct {
	Data any
}

func (J *JSON) Render(w http.ResponseWriter) error {
	J.WriteContentType(w)
	jsonData, err := json.Marshal(J.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonData)
	return err
}

func (J *JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/json; charset=utf-8")
}
