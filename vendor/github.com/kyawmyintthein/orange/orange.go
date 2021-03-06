package orange

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sync"
	"os"
	"runtime"
	"path/filepath"
)

const(
	ConfigFilename = "application"
	ConfigFiletype = "yaml"
)

const(
	ConfigKeyApp = "app"
	ConfigkeyAppName = ConfigKeyApp + ".name"
	ConfigKeyAppEnv  = ConfigKeyApp + ".env"
	ConfigKeyAppEnvs = ConfigKeyApp + ".envs"
)
// buffer pool
var bufPool = newBufferPool(100)

type App struct {
	name       string
	version    string
	rootDir    string
	env        string
	envs       []string
	router     *Router
	httprouter *httprouter.Router
	config     *Config      
	pool       sync.Pool
}

type HandlerFunc func(ctx *Context)

// New: init new app object
func NewApp(name string) *App {
	var app *App
	app = new(App)


	// get caller 
	 _, file, _, _ := runtime.Caller(1)
	fpath := filepath.FromSlash(file)
	dir, file := filepath.Split(fpath)

	app.name = name
	app.rootDir = dir
	app.defaultPool()
	app.newRouter()
	app.loadConfig()
	app.defaultConfig()
	return app
}

// NewConfig: 
func (app *App) NewConfig(filename, path, filetype string) (*Config, error){
	var(
		config *Config
		err     error
	)
	config = new(Config)
	config.filetype = filetype
	config.filename = filename
	if path == ""{
	 	_, fpath, _, _ := runtime.Caller(1)
		config.path = fpath
	}
	config.app = app
	config.replacer = defaultReplacer
	err = config.load()
	return config, err
}

func (app *App) Config(filename, path, filetype string) error{
	var err error
	app.config = new(Config)
	app.config.filetype = filetype
	app.config.filename = filename
	app.config.path = path
	err = app.config.load()
	return err
}

// defaultConfig 
func (app *App) defaultConfig(){
	app.envs = app.config.GetStringSlice(ConfigKeyAppEnvs)
	app.env = app.config.GetString(ConfigKeyAppEnv)
	app.name = app.config.GetString(ConfigkeyAppName)
}

// loadConfig 
func (app *App) loadConfig(){
	var (
		config 	  *Config
		err       error
	)
	config = new(Config)
	config.filename = ConfigFilename
	config.filetype = ConfigFiletype
	config.path = app.rootDir
	if err = config.load(); err != nil{
		colorLog("[ERRO] unable to find the default config file under " + app.rootDir + ".\n")
		os.Exit(1)
	}
	app.config = config
}

// defaultPool: load default pool
func (app *App) defaultPool() {
	app.pool.New = func() interface{} {
		return &Context{app: app, index: -1, data: nil}
	}
}

// newContext: init new Context for each request and response
func (app *App) newContext(rw http.ResponseWriter, req *http.Request) *Context {
	var ctx *Context
	ctx = app.pool.Get().(*Context)
	ctx.request = req
	ctx.response = &Response{app: app}
	ctx.Writer = ctx.response
	ctx.index = -1
	ctx.data = nil
	ctx.response.reset(rw)
	ctx.app = app
	return ctx
}

// newRouter: init new router
func (app *App) newRouter() {
	var (
		hrouter *httprouter.Router
	)
	hrouter = httprouter.New()
	app.router = &Router{
		handlerFuncs: nil,
		prefix:   "/",
		app:      app,
	}
	app.httprouter = hrouter
	app.handleNotFound()
	app.handlePanic()
}

// handleNotFound:  hanlder function for not found
func (app *App) handleNotFound() {
	app.httprouter.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var ctx *Context
		ctx = app.newContext(rw, req)
		ctx.response.Header().Set(HeaderContentType, MIMETypeApplicationJSONCharsetUTF8)
		ctx.response.WriteHeader(http.StatusNotFound)
		ctx.Next()
		b, _ := json.Marshal(notFoundError)
		ctx.response.Write(b)
		app.pool.Put(ctx)
	})
}

// handlePanic: handler function for panic
func (app *App) handlePanic() {
	app.httprouter.PanicHandler = func(rw http.ResponseWriter,req *http.Request,i interface {}){
		var ctx *Context
		ctx = app.newContext(rw, req)
		ctx.response.Header().Set(HeaderContentType, MIMETypeApplicationJSONCharsetUTF8)
		ctx.response.WriteHeader(http.StatusInternalServerError)
		ctx.Next()
		b, _ := json.Marshal(i)
		ctx.response.Write(b)
		app.pool.Put(ctx)
	}
}

// Start: start http server
func (app *App) Start(addr string) {
	colorLog("[INFO] server start at: %s\n", addr)
	if err := http.ListenAndServe(addr, app.router); err != nil {
		panic(err)
	}
}

// Start lts (https) server
func (app *App) StartTLS(addr string, cert string, key string) {
	if err := http.ListenAndServeTLS(addr, cert, key, app.router); err != nil {
		panic(err)
	}
}

// Namespace: add new group router
func (app *App) Namespace(path string) *Router {
	router := Router{
		handlerFuncs: nil,
		prefix:       app.router.path(path),
		app:          app,
	}
	return &router
}

// Set BufferPoolSize
func (a *App) SetPoolSize(poolSize int) {
	bufPool = newBufferPool(poolSize)
}

// Use: use middlewares
func (app *App) Use(middlewares ...HandlerFunc) {
	app.router.handlerFuncs = append(app.router.handlerFuncs, middlewares...)
}

// AppConfig: load config
func (app *App) AppConfig() *Config{
	return app.config
}


// AppConfig: load config
func (app *App) ENV() string{
	return app.env
}

// Version: get version
func (app *App) Version() string{
	return app.version
}

