package defaults

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/infinytum/go-mojito"
	"github.com/infinytum/go-mojito/log"
	"github.com/infinytum/go-mojito/routing"
	"github.com/uptrace/bunrouter"
)

type bunRouter struct {
	defaultHandler mojito.Handler
	errorHandler   mojito.Handler
	middlewares    []interface{}
	router         *bunrouter.CompatRouter
	routeMap       map[string]mojito.Handler
	sync.Mutex
	http.Server
}

//// Convenience functions for registering routes

func (r *bunRouter) CONNECT(path string, handler interface{}) error {
	return r.WithRoute(http.MethodConnect, path, handler)
}

func (r *bunRouter) DELETE(path string, handler interface{}) error {
	return r.WithRoute(http.MethodDelete, path, handler)
}

func (r *bunRouter) GET(path string, handler interface{}) error {
	return r.WithRoute(http.MethodGet, path, handler)
}

func (r *bunRouter) HEAD(path string, handler interface{}) error {
	return r.WithRoute(http.MethodHead, path, handler)
}

func (r *bunRouter) POST(path string, handler interface{}) error {
	return r.WithRoute(http.MethodPost, path, handler)
}

func (r *bunRouter) PUT(path string, handler interface{}) error {
	return r.WithRoute(http.MethodPut, path, handler)
}

func (r *bunRouter) TRACE(path string, handler interface{}) error {
	return r.WithRoute(http.MethodTrace, path, handler)
}

func (r *bunRouter) OPTIONS(path string, handler interface{}) error {
	return r.WithRoute(http.MethodOptions, path, handler)
}

func (r *bunRouter) PATCH(path string, handler interface{}) error {
	return r.WithRoute(http.MethodPatch, path, handler)
}

//// Generic functions for adding routes and middleware

// WithDefaultHandler will set the default handler for the router
func (r *bunRouter) WithDefaultHandler(handler interface{}) error {
	r.Lock()
	defer r.Unlock()
	h, err := routing.NewHandler(handler)
	if err != nil {
		return err
	}
	r.defaultHandler = h
	return nil
}

// WithErrorHandler will set the error handler for the router
func (r *bunRouter) WithErrorHandler(handler interface{}) error {
	r.Lock()
	defer r.Unlock()
	h, err := routing.NewHandler(handler)
	if err != nil {
		return err
	}
	r.errorHandler = h
	return nil
}

// WithGroup will create a new route group for the given prefix
func (r *bunRouter) WithGroup(path string, callback func(group mojito.RouteGroup)) error {
	rg := routing.NewRouteGroup()
	callback(rg)
	return rg.ApplyToRouter(r, path)
}

// WithMiddleware will add a middleware to the router
func (r *bunRouter) WithMiddleware(handler interface{}) error {
	r.Lock()
	defer r.Unlock()
	for _, h := range r.routeMap {
		if err := h.AddMiddleware(handler); err != nil {
			log.Error(err)
			return err
		}
	}
	r.middlewares = append(r.middlewares, handler)
	return nil
}

// WithRoute will add a new route with the given RouteMethod to the router
func (r *bunRouter) WithRoute(method string, path string, handler interface{}) error {
	r.Lock()
	defer r.Unlock()

	// If the handler is already a mojito handler, skip the handler creation
	h, err := routing.GetHandler(handler)
	if err != nil {
		// Check if the handler is of kind func, else this is not a valid handler.
		if reflect.TypeOf(handler).Kind() != reflect.Func {
			return errors.New("The given route handler is neither a func nor a mojito.Handler and is therefore not valid")
		}

		// The handler is of kind func, attempt to create a new mojito.Handler for it
		h, err = routing.NewHandler(handler)
		if err != nil {
			log.Field("method", method).Field("path", path).Errorf("Failed to create a new mojito.Handler for given route handler: %s", err)
			return err
		}
	}

	// Safety check, this should never happen
	if h == nil {
		return errors.New("mojito.Handler was unexpectedly nil")
	}

	// Chain router-wide middleware to the (new) handler
	for _, middleware := range r.middlewares {
		if err := h.AddMiddleware(middleware); err != nil {
			log.Field("method", method).Field("path", path).Errorf("Failed to chain middleware to route: %s", err)
			return err
		}
	}

	switch method {
	case http.MethodDelete:
		r.router.Group.Compat().DELETE(path, r.withMojitoHandler(h))
	case http.MethodGet:
		r.router.Group.Compat().GET(path, r.withMojitoHandler(h))
	case http.MethodHead:
		r.router.Group.Compat().HEAD(path, r.withMojitoHandler(h))
	case http.MethodPost:
		r.router.Group.Compat().POST(path, r.withMojitoHandler(h))
	case http.MethodPut:
		r.router.Group.Compat().PUT(path, r.withMojitoHandler(h))
	default:
		log.Field("method", method).Field("path", path).Error("The default bun router implementation unfortunately does not support this HTTP method")
		return errors.New("The given HTTP method is not available on this router")
	}
	r.routeMap[path] = h
	return nil
}

func (r *bunRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			if r.errorHandler == nil {
				log.Errorf("%s: %s", err, string(debug.Stack()))
			} else {
				req := routing.NewRequest(req)
				res := routing.NewResponse(w)
				res.ViewBag().Set("error", err)
				r.errorHandler.Serve(req, res)
			}
		}
	}()
	r.router.ServeHTTP(w, req)
}

func (r *bunRouter) ServeNotFound(w http.ResponseWriter, re *http.Request) {
	if r.defaultHandler == nil {
		return
	}
	req := routing.NewRequest(re)
	res := routing.NewResponse(w)
	r.defaultHandler.Serve(req, res)
}

// ListenAndServe will start an HTTP webserver on the given address
func (r *bunRouter) ListenAndServe(address string) error {
	r.Server = http.Server{
		Addr:    address,
		Handler: http.HandlerFunc(r.ServeHTTP),
	}
	return r.Server.ListenAndServe()
}

// Shutdown will stop the HTTP webserver
func (r *bunRouter) Shutdown() error {
	return r.Server.Shutdown(context.TODO())
}

//// Internal functions

func (r *bunRouter) withMojitoHandler(handler mojito.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := routing.NewRequest(r)
		res := routing.NewResponse(w)
		req.SetParams(bunrouter.ParamsFromContext(r.Context()).Map())
		if err := handler.Serve(req, res); err != nil {
			panic(err)
		}
	}
}

// NewBunRouter will create new instance of the mojito bun router implementation
func newBunRouter() mojito.Router {
	bunRouter := &bunRouter{
		routeMap: make(map[string]mojito.Handler),
		Mutex:    sync.Mutex{},
	}
	bunRouter.router = bunrouter.New(bunrouter.WithNotFoundHandler(bunrouter.HTTPHandlerFunc(bunRouter.ServeNotFound))).Compat()
	return bunRouter
}
