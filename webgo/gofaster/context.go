package gofaster

import (
	"gofaster/render"
	"net/http"
	"net/url"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	e *Engine
}

func (c *Context) HTML(status int, html string) error {
	return c.Render(status, &render.HTML{Data: html, IsTemplate: false})
}

func (c *Context) Template(name string, data any) error {
	//状态为200
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		Name:       name,
		Template:   c.e.HTMLRender.Template,
		IsTemplate: true,
	})
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(status, &render.JSON{Data: data})
}

func (c *Context) XML(status int, data any) error {
	return c.Render(status, &render.XML{Data: data})
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

func (c *Context) Redirect(status int, location string) error {
	//status 需为30*
	return c.Render(status, &render.Redirect{
		Code:     status,
		Location: location,
		Request:  c.R,
	})
}
func (c *Context) String(status int, format string, values ...any) error {
	return c.Render(status, &render.String{
		Format: format,
		Data:   values,
	})
}

func (c *Context) Render(status int, r render.Render) error {
	c.W.WriteHeader(status)
	return r.Render(c.W)
}
