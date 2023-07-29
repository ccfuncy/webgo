package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type FsHttpClient struct {
	client     http.Client
	ServiceMap map[string]FsService
}

func (c *FsHttpClient) RegisterHttpServiceName(name string, service FsService) {
	c.ServiceMap[name] = service
}

func NewFsHttpClient() *FsHttpClient {
	//transport 请求分发，协程安全的
	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Duration(3) * time.Second,
	}
	return &FsHttpClient{client: client, ServiceMap: make(map[string]FsService)}
}
func (c *FsHttpClient) Get(url string, args map[string]any) ([]byte, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + c.toValues(args)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FsHttpClient) PostForm(url string, args map[string]any) ([]byte, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FsHttpClient) PostJson(url string, args map[string]any) ([]byte, error) {
	marshal, _ := json.Marshal(args)
	request, err := http.NewRequest("POST", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FsHttpClient) GetRequest(url string, args map[string]any) (*http.Request, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + c.toValues(args)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *FsHttpClient) FormRequest(url string, args map[string]any) (*http.Request, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *FsHttpClient) JsonRequest(url string, args map[string]any) (*http.Request, error) {
	marshal, _ := json.Marshal(args)
	request, err := http.NewRequest("POST", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *FsHttpClient) Response(r *http.Request) ([]byte, error) {
	return c.responseHandle(r)
}

func (c *FsHttpClient) responseHandle(request *http.Request) ([]byte, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		info := fmt.Sprintf("response status is %d", response.StatusCode)
		return nil, errors.New(info)
	}
	reader := bufio.NewReader(response.Body)
	var buf []byte = make([]byte, 127)
	var body []byte
	for true {
		n, err := reader.Read(buf)
		if err == io.EOF || n == 0 {
			break
		}
		body = append(body, buf[:n]...)
		if n < len(buf) {
			break
		}
	}
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *FsHttpClient) toValues(args map[string]any) string {
	values := url.Values{}
	for k, v := range args {
		values.Set(k, fmt.Sprintf("%v", v))
	}
	return values.Encode()
}

type HttpConfig struct {
	Protocol string
	Host     string
	Port     int
}

const (
	HTTP  = "http"
	HTTPS = "https"
)
const (
	GET      = "GET"
	POSTForm = "POST_FORM"
	POSTJson = "POST_JSON"
)

type FsService interface {
	Env() *HttpConfig
}

func (c *FsHttpClient) Do(service string, method string) FsService {
	fsService, ok := c.ServiceMap[service]
	if !ok {
		panic(errors.New("service not register"))
	}
	//反射，取值
	v := reflect.ValueOf(fsService)
	t := reflect.TypeOf(fsService)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("service not pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	var methodIndex = -1
	for i := 0; i < tVar.NumField(); i++ {
		if tVar.Field(i).Name == method {
			methodIndex = i
			break
		}
	}
	if methodIndex == -1 {
		panic(errors.New("method not found"))
	}
	tag := tVar.Field(methodIndex).Tag
	rpcInfo := tag.Get("fsrpc")
	if rpcInfo == "" {
		panic(errors.New("fsrpc info not found"))
	}
	split := strings.Split(rpcInfo, ",")
	if len(split) != 2 {
		panic(errors.New("tag fsrpc not vaild"))
	}
	methodType := split[0]
	path := split[1]
	httpConfig := fsService.Env()

	f := func(args map[string]any) ([]byte, error) {
		switch methodType {
		case GET:
			return c.Get(httpConfig.Prefix()+path, args)
		case POSTJson:
			return c.PostJson(httpConfig.Prefix()+path, args)
		case POSTForm:
			return c.PostForm(httpConfig.Prefix()+path, args)
		}
		return nil, errors.New("no match method type")
	}
	value := reflect.ValueOf(f)
	vVar.Field(methodIndex).Set(value)
	return fsService
}
func (c HttpConfig) Prefix() string {
	if c.Protocol == "" {
		c.Protocol = HTTP
	}
	switch c.Protocol {
	case HTTP:
		return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
	case HTTPS:
		return fmt.Sprintf("https://%s:%d", c.Host, c.Port)
	}
	return ""
}
