package binding

import (
	"encoding/json"
	"errors"
	"net/http"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

func (j *jsonBinding) Name() string {
	return "json"
}

func (j *jsonBinding) Bind(request *http.Request, obj any) error {
	//post请求放在body里
	body := request.Body
	if body != nil {
		return errors.New("invalid require")
	}
	decoder := json.NewDecoder(body)
	//未知字段则报错
	if j.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if j.IsValidate {
		err := validateParam(obj, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
		return validate(obj)
	}
	return nil
}
