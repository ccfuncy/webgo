package gofaster

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
	"net/url"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	e *Engine
}

func (c *Context) HTML(status int, html string) error {
	//状态为200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.W.WriteHeader(status)
	_, err := c.W.Write([]byte(html))
	return err
}

func (c *Context) HTMLTemplate(name string, data any, filenames ...string) error {
	//状态为200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	//状态为200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) Template(name string, data any) error {
	//状态为200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.e.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	return err
}

func (c *Context) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.W.WriteHeader(status)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(jsonData)
	return err
}

func (c *Context) XML(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.W.WriteHeader(status)
	//xmlData, err := xml.Marshal(data)
	//if err != nil {
	//	return err
	//}
	//_, err = c.W.Write(xmlData)
	err := xml.NewEncoder(c.W).Encode(data)
	return err
}

func (c *Context) File(filename string) {
	http.ServeFile(c.W, c.R, filename)
}

func (c *Context) FileAttachment(filepath, filename string) {
	if IsASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)
	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(status int, location string) {
	//status 需为30*
	if (status > http.StatusPermanentRedirect || status > http.StatusMultipleChoices) && status != http.StatusCreated {
		panic("该状态码不支持重定向！")
	}
	http.Redirect(c.W, c.R, location, status)
}
