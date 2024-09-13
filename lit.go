package lit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Map map[string]any

type HandlerFunc func(c *Context) error

type ErrHandlerFunc func(err error, c *Context)

func handleError(err error, c *Context) {
	// if c.Res.Commited {
	// if the response has already been committed, log the error and return
	// this is a safety measure to prevent writing to a closed connection
	// also prevents the error from being written to the response but it is
	// not handled at the app level it will be lost unless we handle it here.
	//
	// just log it here for now
	// fmt.Printf("error in committed response: %v\n", err)
	// return
	// }

	var httpErr *HTTPError

	if errors.As(err, &httpErr) {
		errors.As(httpErr.Internal, &httpErr)
	} else {
		log.Println(err)
		httpErr = ErrInternalServerError
	}

	err = c.JSON(httpErr, httpErr.Code)

	if err != nil {
		log.Println(err)
	}
}

func handleNotFound(c *Context) error {
	return ErrNotFound
}

type App struct {
	mux             *http.ServeMux
	middleware      []HandlerFunc
	ErrorHandler    ErrHandlerFunc
	NotFoundHandler HandlerFunc
}

func New() *App {
	return &App{
		mux:             http.NewServeMux(),
		ErrorHandler:    handleError,
		NotFoundHandler: handleNotFound,
	}
}

// Entry point for the request of the application.
// It is responsible of creating the context and passing it to the middleware chain.
// It also handles the response error and calls the error handler.
//
// It should be handled by the App.ServeHTTP method.
// but for now this is the top most layer that i can handle it from passing the context
// to the middleware and handling the error.
func (a *App) NewHandler(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			app:    a,
			Req:    r,
			Res:    NewResponse(w),
			Query:  r.URL.Query(),
			Header: w.Header(),
			Next:   func() error { return nil },
		}

		if err := h(c); err != nil {
			a.ErrorHandler(err, c)
		}
	})
}

// Creates a chain of handlers to be executed in reverse order.
func compose(handlers []HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		i := len(handlers)
		c.Next = func() error {
			i--
			if i < 0 {
				return ErrNoNextHandler
			}
			return handlers[i](c)
		}

		return c.Next()
	}
}

func (a *App) Use(h HandlerFunc) {
	a.middleware = append(a.middleware, h)
}

func (a *App) Add(method, path string, h HandlerFunc, hs ...HandlerFunc) {
	chain := append(append(hs, h), a.middleware...)
	h = compose(chain)
	// addRoute(a.mux, method, FmtPath(path), h, a.ErrorHandler)
	a.mux.Handle(Pattern(path, method), a.NewHandler(h))
}

func (a *App) GET(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodGet, path, h, hs...)
}

func (a *App) POST(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodPost, path, h, hs...)
}

func (a *App) PUT(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodPut, path, h, hs...)
}

func (a *App) PATCH(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodPatch, path, h, hs...)
}

func (a *App) DELETE(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodDelete, path, h, hs...)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)

	// there must be a way handle the logic here instead of in the mux
	// c := &Context{
	// 	Req:    r,
	// 	Res:    NewResponse(w),
	// 	Query:  r.URL.Query(),
	// 	Header: w.Header(),
	// 	Next:   func() error { return nil },
	// 	app:    a,
	// }
	//
	// _, pattern := a.mux.Handler(r)
}

func Encode[T any](w http.ResponseWriter, r *http.Request, v T) error {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func GetHeader(r *http.Request, key string) (string, bool) {
	v := r.Header.Get(key)
	return v, v != ""
}

// Response implements the ResponseWriter interface, it is a thin wrapper
// to have more control over the response.
//
// A ResponseWriter interface is used by an HTTP handler to
// construct an HTTP response.
type Response struct {
	http.ResponseWriter
	Size     int64
	Status   int
	Commited bool
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{ResponseWriter: w}
}

func (r *Response) WriteHeader(code int) {
	if r.Commited {
		log.Println("response already committed")
		return
	}

	r.Status = code
	r.ResponseWriter.WriteHeader(code)
	r.Commited = true
}

func (r *Response) Write(b []byte) (int, error) {
	if !r.Commited {
		if r.Status == 0 {
			r.Status = http.StatusOK
		}
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.ResponseWriter.Write(b)
	r.Size += int64(n)
	return n, err
}

// Unwrap returns the original http.ResponseWriter.
// ResponseController can be used to access the original http.ResponseWriter.
// See [https://go.dev/blog/go1.20]
func (r *Response) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func FmtStatus(code int) string {
	// print color based on status code range
	switch {
	case code >= 200 && code < 300:
		return FmtColor(code, "green")
	case code >= 300 && code < 400:
		return FmtColor(code, "cyan")
	case code >= 400 && code < 500:
		return FmtColor(code, "white")
	case code >= 500:
		return FmtColor(code, "red")
	default:
		return string(code)
	}
}

func IsRoot(path string) bool {
	return path == "/"
}

func FmtPath(path string) string {
	if IsRoot(path) {
		return "/{$}"
	}
	return path
}

func Pattern(path, method string) string {
	return method + " " + FmtPath(path)
}
