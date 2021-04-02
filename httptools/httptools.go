package httptools

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"

	"github.com/julienschmidt/httprouter"
)

type HTTPError struct {
	err error
	status int
}

type RouteHandle func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error)
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
			status, err := h(w, r, ps)
			if err != nil {
				http.Error(w, fmt.Sprintf("steno: %s", err.Error()), status)
				log.Printf("ERROR/handler/%s :%s\n",
					runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
					err.Error())
				break
			}
		}
	}
}

func (rt Route) Finish(handler RouteHandle) httprouter.Handle {
	return rt.Gate(handler).Handle()
}

func (rt Route) Apply(handler httprouter.Handle) Route {
	rt.handlers = append(rt.handlers,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
			handler(w, r, ps)
			return http.StatusOK, nil
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
