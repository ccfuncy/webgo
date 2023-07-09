package render

import (
	"errors"
	"net/http"
)

type Redirect struct {
	Code     int
	Location string
	Request  *http.Request
}

func (r *Redirect) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if (r.Code > http.StatusPermanentRedirect ||
		r.Code > http.StatusMultipleChoices) &&
		r.Code != http.StatusCreated {
		return errors.New("该Code不支持重定向")
	}
	http.Redirect(w, r.Request, r.Location, r.Code)
	return nil
}

func (r *Redirect) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
