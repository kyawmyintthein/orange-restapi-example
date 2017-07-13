package orange

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type ContentType string
type Charset string

const (
	contentType    = "Content-Type"
	acceptLanguage = "Accept-Language"
	abortIndex     = math.MaxInt8 / 2
	httpProtocol   = "http"
	httpsProtocol  = "https"
	defaultMemory  = 32 << 20 // 32 MB
)

const (
	ContentTypeJSON ContentType = "applicaiton/json"
)

const (
	charsetUTF8 string = "UIF-8"
)

type Context interface {
	// Request returns *Request
	Request() *http.Request

	// Response returns *Response.
	Response() *Response

	// Scheme returns http protocol scheme
	Scheme() string

	// ClientIP returns client IP address
	ClientIP() string

	// Path returns url path
	Path() string

	// Param returns a param
	Param(name string) interface{}

	// Params return all params
	Params() map[string]interface{}

	// QueryParam return a query string param
	QueryParam(name string) url.Values

	// QueryParams return all query string params
	QueryParams() map[string]interface{}

	// QueryString return query string
	QueryString() string

	FormValue(name string) interface{}

	FormData() (url.Values, error)

	File(name string) (*multipart.FileHeader, error)

	MultipartForm() (*multipart.Form, error)

	Cookie(name string) (*http.Cookie, error)

	SecureCookie(name string) (*http.Cookie, error)

	Cookies() []*http.Cookie

	SetCookie(name, value string, others ...interface{})

	SetSecureCookie(Secret, name, value string, others ...interface{})

	ResponseJSON(status int, data interface{})

	ResponseJSONP(status int, callback string, data interface{})

	ResponseBytes(status int, contentType string, data []byte)

	App() *App

	Next()

	Abort()
}

type context struct {
	response     *Response
	request      *http.Request
	query        url.Values
	params       httprouter.Params
	path         string
	data         map[string]interface{}
	app          *App
	handlerFuncs []HandlerFunc
	index        int8
}

// Request: returns request
func (ctx *context) Request() *http.Request {
	return ctx.request
}

// Response: return response
func (ctx *context) Response() *Response {
	return ctx.response
}

// Scheme: return http protocol schame as string
func (ctx *context) Scheme() string {
	if ctx.IsTLS() {
		return httpsProtocol
	}
	if scheme := ctx.request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := ctx.request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := ctx.request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := ctx.request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

// ClientIP: return ip address of client
func (ctx *context) ClientIP() string {
	var (
		remoteAddress, ip string
	)
	remoteAddress = ctx.request.RemoteAddr
	if ip = ctx.request.Header.Get(HeaderXForwardedFor); ip != "" {
		remoteAddress = strings.Split(ip, ", ")[0]
	} else if ip = ctx.request.Header.Get(HeaderXRealIP); ip != "" {
		remoteAddress = ip
	} else {
		remoteAddress, _, _ = net.SplitHostPort(remoteAddress)
	}
	return remoteAddress
}

// Path: return url path
func (ctx *context) Path() string {
	return ctx.path
}

// ResponseJSON: response json to client
func (ctx *context) ResponseJSON(status int, data interface{}) {
	ctx.response.Header().Set(contentType, fmt.Sprintf("%s; charset=%s", ContentTypeJSON, charsetUTF8))
	ctx.response.status = status
	ctx.response.WriteHeader()
	if data == nil {
		return
	}
	b, _ := json.Marshal(data)
	ctx.response.Write([]byte(b))
}

// ResponseJSONP: response jsonp to client
func (ctx *context) ResponseJSONP(status int, callback string, data interface{}) {
	ctx.response.Header().Set(contentType, fmt.Sprintf("%s; charset=%s", ContentTypeJSON, charsetUTF8))
	ctx.response.status = status
	ctx.response.WriteHeader()
	if data == nil {
		return
	}
	datab, _ := json.Marshal(data)
	b := []byte(callback + "(" + string(datab) + ")")
	ctx.response.Write(b)
}

func (ctx *context) ResponseBytes(status int, contentType string, data []byte) {
	ctx.response.Header().Set(HeaderContentType, contentType)

	ctx.response.status = status
	ctx.response.WriteHeader()
	ctx.response.Write(data)
}

// Param: get param from route
func (ctx *context) Param(name string) string {
	return ctx.params.ByName(name)
}

// Param: get all params from httprouter
func (ctx *context) Params() map[string]interface{} {
	var params = make(map[string]interface{})
	for _, param := range ctx.params {
		params[param.Key] = param.Value
	}
	return params
}

// QueryParam: get parameter by name from query string
func (ctx *context) QueryParam(name string) string {
	return ctx.request.URL.Query().Get(name)
}

// QueryParams: get all query string parameters
func (ctx *context) QueryParams() url.Values {
	return ctx.request.URL.Query()
}

// QueryParams: get query string
func (ctx *context) QueryString() string {
	return ctx.request.URL.RawQuery
}

// FormValue: return form value as string
func (ctx *context) FormValue(name string) string {
	return ctx.request.FormValue(name)
}

// FormData: return form values
func (ctx *context) FormData() (url.Values, error) {
	var err error
	if strings.HasPrefix(ctx.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err = ctx.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err = ctx.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return ctx.request.Form, nil
}

func (ctx *context) File(name string) (*multipart.FileHeader, error) {
	_, fileheader, err := ctx.request.FormFile(name)
	return fileheader, err
}

func (ctx *context) MultipartForm() (*multipart.Form, error) {
	err := ctx.request.ParseMultipartForm(defaultMemory)
	return ctx.request.MultipartForm, err
}

func (ctx *context) Cookie(name string) (*http.Cookie, error) {
	return ctx.request.Cookie(name)
}

func (ctx *context) Cookies() []*http.Cookie {
	return ctx.request.Cookies()
}

func (ctx *context) App() *App {
	return ctx.app
}

func (ctx *context) Next() {
	ctx.index++
	s := int8(len(ctx.handlerFuncs))
	for ; ctx.index < s; ctx.index++ {
		ctx.handlerFuncs[ctx.index](ctx)
	}
}

func (ctx *context) Abort() {
	ctx.index = abortIndex
}

func (ctx *context) IsTLS() bool {
	return false
}
