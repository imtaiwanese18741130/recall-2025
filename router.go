package main

import (
	"context"
	"net/http"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func withRecovery(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

func chainMiddleware(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type RouteRegistrar struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewRouteRegistrar(middlewares ...Middleware) *RouteRegistrar {
	return &RouteRegistrar{
		mux:         http.NewServeMux(),
		middlewares: middlewares,
	}
}

func (r RouteRegistrar) GetMux() *http.ServeMux {
	return r.mux
}

func (r *RouteRegistrar) Set404Response(notFound http.HandlerFunc) {
	r.mux.HandleFunc("/", notFound)
}

func (r *RouteRegistrar) handle(method, path string, handler http.HandlerFunc) {
	wrapped := chainMiddleware(handler, append(r.middlewares, withRecovery)...)

	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		wrapped(w, req)
	})
}

func (r *RouteRegistrar) GET(path string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

func (r *RouteRegistrar) POST(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

func (r *RouteRegistrar) PUT(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

func (r *RouteRegistrar) DELETE(path string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

func (r *RouteRegistrar) PATCH(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

func (r *RouteRegistrar) OPTIONS(path string, handler http.HandlerFunc) {
	r.handle(http.MethodOptions, path, handler)
}

type ctxKey string

func SetParam(r *http.Request, key, val string) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKey(key), val)
	return r.WithContext(ctx)
}

func Param(r *http.Request, key string) string {
	if v := r.Context().Value(ctxKey(key)); v != nil {
		return v.(string)
	}
	return ""
}
