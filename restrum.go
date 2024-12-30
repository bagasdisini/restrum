package restrum

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// HandlerFunc defines the handler used by middleware as return value.
type HandlerFunc func(*Context)

// RouterGroup represents a group of routes with a common prefix and middleware.
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

// Engine is the main struct of the framework. It contains the router and configuration.
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
	config Config
}

// Config holds the configuration for the Engine.
type Config struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowCredentials bool
}

// New creates a new Engine instance with optional configuration.
func New(cfg ...Config) *Engine {
	var config Config
	if len(cfg) > 0 {
		config = cfg[0]
	}

	engine := &Engine{
		router: NewRouter(),
		config: config,
	}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{
		engine.RouterGroup,
	}
	return engine
}

// Group creates a new RouterGroup with the given prefix.
func (e *RouterGroup) Group(prefix string) *RouterGroup {
	engine := e.engine
	newGroup := &RouterGroup{
		prefix: e.prefix + prefix,
		parent: e,
		engine: engine,
	}

	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use adds middleware to the RouterGroup.
func (e *RouterGroup) Use(middlewares ...HandlerFunc) {
	e.middlewares = append(e.middlewares, middlewares...)
}

// AddRoutes adds a route to the router with the given method, pattern, and handler.
func (e *RouterGroup) AddRoutes(method string, comp string, handler HandlerFunc) {
	pattern := e.prefix + comp
	e.engine.router.AddRoutes(method, pattern, handler)
}

// GET adds a GET route to the router.
func (e *RouterGroup) GET(pattern string, handler HandlerFunc) {
	e.AddRoutes("GET", pattern, handler)
}

// POST adds a POST route to the router.
func (e *RouterGroup) POST(pattern string, handler HandlerFunc) {
	e.AddRoutes("POST", pattern, handler)
}

// PUT adds a PUT route to the router.
func (e *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	e.AddRoutes("PUT", pattern, handler)
}

// DELETE adds a DELETE route to the router.
func (e *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	e.AddRoutes("DELETE", pattern, handler)
}

// OPTION adds an OPTION route to the router.
func (e *Engine) OPTION(pattern string, handler HandlerFunc) {
	e.router.AddRoutes("OPTION", pattern, handler)
}

// Run starts the HTTP server on the specified address.
func (e *Engine) Run(addr string) (err error) {
	if isPortInUse(addr) {
		panic("port was used!")
	}

	log.Printf("http server running on %s", addr)
	return http.ListenAndServe(addr, e)

}

// ServeHTTP implements the http.Handler interface to handle HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	ctx := newContext(w, req, &e.config)
	ctx.middleware = middlewares
	cfg := &handlerCfg{ctx}
	e.router.handle(cfg)
}

// isPortInUse checks if the specified port is already in use.
func isPortInUse(port string) bool {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return true
	}
	err = ln.Close()
	if err != nil {
		return false
	}
	return false
}

// CORSMiddleware creates a middleware to handle CORS requests.
func CORSMiddleware(config *Config) HandlerFunc {
	return func(ctx *Context) {
		origin := ctx.Request.Header.Get("Origin")

		allowedOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin != "" {
			ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		if len(config.AllowMethods) > 0 {
			ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
		}

		if config.AllowCredentials {
			ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if ctx.Request.Method == "OPTIONS" {
			ctx.ResponseWriter.WriteHeader(http.StatusOK)
			return
		}
		ctx.Next()
	}
}

// joinStrings joins a slice of strings with the specified separator.
func joinStrings(items []string, sep string) string {
	return strings.Join(items, sep)
}
