package restrum

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
)

// Context represents the context of the current HTTP request.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]string
	HTTPMethod     string
	RoutePath      string
	ResponseCode   int

	current    int
	config     *Config
	middleware []HandlerFunc
}

// newContext creates a new Context instance.
func newContext(w http.ResponseWriter, r *http.Request, config *Config) *Context {
	return &Context{
		Request:        r,
		ResponseWriter: w,
		HTTPMethod:     r.Method,
		RoutePath:      r.URL.Path,

		current: -1,
		config:  config,
	}
}

// Next executes the next middleware in the chain.
func (ctx *Context) Next() {
	ctx.current++
	if ctx.current < len(ctx.middleware) {
		ctx.middleware[ctx.current](ctx)
	}
}

// FormValue returns the form value associated with the given key.
func (ctx *Context) FormValue(key string) string {
	return ctx.Request.FormValue(key)
}

// Param returns the URL parameter associated with the given key.
func (ctx *Context) Param(key string) string {
	return ctx.Params[key]
}

// QueryParam returns the query parameter associated with the given key.
func (ctx *Context) QueryParam(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

// String sends a plain text response with the given status code and format.
func (ctx *Context) String(code int, format string) {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/plain")
	ctx.ResponseCode = code
	ctx.ResponseWriter.WriteHeader(code)

	_, err := ctx.ResponseWriter.Write([]byte(format))
	if err != nil {
		return
	}
}

// JSON sends a JSON response with the given status code and object.
func (ctx *Context) JSON(code int, object interface{}) {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	ctx.ResponseCode = code
	ctx.ResponseWriter.WriteHeader(code)

	encode := json.NewEncoder(ctx.ResponseWriter)
	if err := encode.Encode(object); err != nil {
		http.Error(ctx.ResponseWriter, err.Error(), 500)
	}
}

// Data sends a binary data response with the given status code.
func (ctx *Context) Data(code int, data []byte) {
	ctx.ResponseCode = code
	ctx.ResponseWriter.WriteHeader(code)

	_, err := ctx.ResponseWriter.Write(data)
	if err != nil {
		return
	}
}

// HTML sends an HTML response with the given status code and HTML content.
func (ctx *Context) HTML(code int, html string) {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/html")
	ctx.ResponseCode = code
	ctx.ResponseWriter.WriteHeader(code)

	_, err := ctx.ResponseWriter.Write([]byte(html))
	if err != nil {
		return
	}
}

// RenderHTML renders an HTML template with the given name and data.
func (ctx *Context) RenderHTML(name string, data any) error {
	tmpl := template.Must(template.ParseFiles(name))
	return tmpl.Execute(ctx.ResponseWriter, data)
}

// Bind binds the request body to the given object.
func (ctx *Context) Bind(d any) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(ctx.Request.Body)
	return decoder.Decode(d)
}

// SetCookie sets a cookie in the response.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.ResponseWriter, cookie)
}

// GetCookie retrieves a cookie from the request by name.
func (ctx *Context) GetCookie(name string) *string {
	cookie, err := ctx.Request.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(ctx.ResponseWriter, "cookie not found", http.StatusBadRequest)
		} else {
			http.Error(ctx.ResponseWriter, "server error", http.StatusInternalServerError)
		}
		return nil
	}
	return &cookie.Value
}

// DeleteCookie deletes a cookie from the response by name.
func (ctx *Context) DeleteCookie(name string) {
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	})
}

// GetIPAddress retrieves the IP address of the client making the request.
func (ctx *Context) GetIPAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "Error getting network interfaces: %v\n", err)
		if err != nil {
			return ""
		}
		os.Exit(1)
	}

	for _, i := range interfaces {
		adds, err := i.Addrs()
		if err != nil {
			_, err = fmt.Fprintf(os.Stderr, "Error getting addresses for interface %v: %v\n", i.Name, err)
			if err != nil {
				return ""
			}
			continue
		}

		for _, addr := range adds {
			ip := extractIP(addr)
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String()
			}
		}
	}

	_, err = fmt.Fprintln(os.Stderr, "No valid IP address found.")
	if err != nil {
		return ""
	}
	os.Exit(1)
	return ""
}

// extractIP extracts the IP address from a net.Addr.
func extractIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		return nil
	}
}
