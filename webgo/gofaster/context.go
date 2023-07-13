package gofaster

import (
	"errors"
	"gofaster/binding"
	"gofaster/render"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMaxMemory = 32 << 20 //32M

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
	e                     *Engine
	StatusCode            int
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}
func (c *Context) get(cache map[string][]string, key string) (map[string]string, bool) {
	dict := make(map[string]string)
	exist := false
	//user[name]=12&&user[id]=1 key user[name],
	for key, value := range cache {
		if i := strings.IndexByte(key, '['); i >= 1 && key[0:i] == key {
			if j := strings.IndexByte(key[i+1:], ']'); j >= 1 {
				exist = true
				dict[key[i+1:][:j]] = value[0]
			}
		}
	}
	return dict, exist
}

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}

}
func (c *Context) initFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMaxMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	} else {
		c.queryCache = url.Values{}
	}
}
func (c *Context) GetForm(key string) string {
	c.initFormCache()
	return c.formCache.Get(key)
}

func (c *Context) GetFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}
func (c *Context) GetFormArray(key string) ([]string, bool) {
	c.initFormCache()
	values, ok := c.formCache[key]
	return values, ok
}

func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMaxMemory)
	return c.R.MultipartForm, err
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
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
	c.StatusCode = status
	if status != http.StatusOK {
		c.W.WriteHeader(status)
	}
	return r.Render(c.W)
}

func (c *Context) bindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

func (c *Context) bindJson(obj any) error {
	json := binding.JSON
	json.DisallowUnknownFields = true
	json.IsValidate = true
	return c.MustBindWith(obj, json)
}

func (c *Context) MustBindWith(obj any, bind binding.Binding) error {
	if err := bind.Bind(c.R, obj); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}
