package restrum

import "net/http"

// router represents the routing tree and handlers.
type router struct {
	root     map[string]*node
	handlers map[string]HandlerFunc
}

// handlerCfg holds the context for the handler.
type handlerCfg struct {
	Ctx *Context
}

// NewRouter creates a new router instance.
func NewRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
		root:     make(map[string]*node),
	}
}

// AddRoutes adds a route to the router with the given method, pattern, and handler.
func (r *router) AddRoutes(method, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "_" + pattern

	if _, ok := r.root[method]; !ok {
		r.root[method] = &node{}
	}

	r.root[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// getRoute retrieves the node and parameters for the given method and path.
func (r *router) getRoute(method, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.root[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for i, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[i]
			} else if part[0] == '*' {
				params[part[1:]] = joinParts(searchParts[i:])
				break
			}
		}
		return n, params
	}
	return nil, nil
}

// handle processes the request and executes the corresponding handler.
func (r *router) handle(ctx *handlerCfg) {
	n, params := r.getRoute(ctx.Ctx.HTTPMethod, ctx.Ctx.RoutePath)
	if n != nil {
		ctx.Ctx.Params = params
		key := ctx.Ctx.HTTPMethod + "_" + n.pattern
		r.handlers[key](ctx.Ctx)
	} else {
		http.Error(ctx.Ctx.ResponseWriter, "NOT FOUND", http.StatusNotFound)
	}
}

// joinParts joins a slice of parts into a single string with '/' separator.
func joinParts(parts []string) string {
	length := 0
	for _, part := range parts {
		length += len(part) + 1
	}
	joined := make([]byte, length-1)
	pos := 0
	for _, part := range parts {
		copy(joined[pos:], part)
		pos += len(part)
		if pos < length-1 {
			joined[pos] = '/'
			pos++
		}
	}
	return string(joined)
}
