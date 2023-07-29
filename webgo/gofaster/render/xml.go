package render

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data any
}

func (x *XML) Render(w http.ResponseWriter) error {
	x.WriteContentType(w)
	//w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	//xmlData, err := xml.Marshal(data)
	//if err != nil {
	//	return err
	//}
	//_, err = c.W.Write(xmlData)
	err := xml.NewEncoder(w).Encode(x.Data)
	return err
}

func (x *XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/xml; charset=utf-8")
}
