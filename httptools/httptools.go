package httptools

import (
	"net/http"
	"log"

	"github.com/julienschmidt/httprouter"
)

type RouteHandle func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool
type Route struct {
	handlers []RouteHandle
}

func RouteNew() Route {
	r := Route{
		handlers: make([]RouteHandle, 0, 4),
	}
	return r
}

func (rt Route) Clone() Route {
	newRoute := RouteNew()
	newRoute.handlers = make([]RouteHandle, len(rt.handlers))
	copy(newRoute.handlers, rt.handlers)

	return newRoute
}

func (rt Route) Handle() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		for _, h := range rt.handlers {
			if !h(w, r, ps) {
				break
			}
		}
	}
}

func (rt Route) Finish(handler httprouter.Handle) httprouter.Handle {
	return rt.Apply(handler).Handle()
}

func (rt Route) Apply(handler httprouter.Handle) Route {
	rt.handlers = append(rt.handlers,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool {
			handler(w, r, ps)
			return true
		})
	return rt
}

func (rt Route) Gate(handler RouteHandle) Route {
	rt.handlers = append(rt.handlers, handler)
	return rt
}

func (rt Route) Log() Route {
	return rt.Apply(httplog)
}

func httplog(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Printf("%s %s --- %s %s", r.UserAgent(), r.RemoteAddr, r.Method, r.URL)
}
