package lit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Map map[string]any

type HandlerFunc func(c *Context) error

type ErrHandlerFunc func(err error, c *Context)

func handleError(err error, c *Context) {
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

type App struct {
	mux          *http.ServeMux
	middleware   []HandlerFunc
	ErrorHandler ErrHandlerFunc
}

func New() *App {
	return &App{
		mux:          http.NewServeMux(),
		ErrorHandler: handleError,
	}
}

func addRoute(mux *http.ServeMux, method, path string, h HandlerFunc, eh ErrHandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}

	mux.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			Req:   r,
			Res:   NewResponse(w),
			Query: r.URL.Query(),
			Next:  func() error { return nil },
		}

		if err := h(c); err != nil {
			eh(err, c)
		}

	})
}

func compose(handlers []HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		fmt.Println("compose", len(handlers))
		i := len(handlers)
		c.Next = func() error {
			i--
			fmt.Println("next", i)
			if i < 0 {
				return errors.New("next called after the last handler")
			}
			fmt.Println("call", i)
			return handlers[i](c)
		}

		return c.Next()
	}
}

func (a *App) Add(method, path string, handlers ...HandlerFunc) {
	h := compose(append(handlers, a.middleware...))
	addRoute(a.mux, method, path, h, a.ErrorHandler)
}

func (a *App) Use(h HandlerFunc) {
	a.middleware = append(a.middleware, h)
}

func (a *App) GET(path string, h HandlerFunc, hs ...HandlerFunc) {
	a.Add(http.MethodGet, path, append(hs, h)...)
}

func (a *App) POST(path string, h HandlerFunc) {
	a.Add(http.MethodPost, path, h)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)

	// there must be a way to create the context here and pass it to the handler
	/* h, p := a.mux.Handler(r)

	c := &Context{
		Req:   r,
		Res:   NewResponse(w),
		Query: r.URL.Query(),
	} */

}

type Context struct {
	Req   *http.Request
	Res   *Response
	Query url.Values
	Next  func() error
}

func (c *Context) writeHeader(code []int) {
	if len(code) > 0 {
		c.Res.WriteHeader(code[0])
	}
}

func (c *Context) Body(b []byte, ct string, code ...int) error {
	c.Res.Header().Set(HeaderContentType, ct)
	c.writeHeader(code)

	_, err := c.Res.Write(b)
	return err
}

func (c *Context) Text(s string, code ...int) error {
	return c.Body([]byte(s), MIMETextPlain, code...)
}

func (c *Context) HTML(s string, code ...int) error {
	return c.Body([]byte(s), MIMETextHTML, code...)
}

func (c *Context) JSON(v any, code ...int) error {
	c.Res.Header().Set(HeaderContentType, MIMEApplicationJSON)
	c.writeHeader(code)

	return Encode(c.Res, c.Req, v)
}

func (c *Context) Path() string {
	return c.Req.URL.Path
}

func (c *Context) Param(name string) string {
	return c.Req.PathValue(name)
}

func (c *Context) NotFound() error {
	return ErrNotFound
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
	Status   int
	Commited bool
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{ResponseWriter: w}
}

func RequestLogger() HandlerFunc {
	return func(c *Context) error {
		start := time.Now()
		m := FmtColor(c.Req.Method, "green")
		u := FmtColor(c.Req.URL.String(), "cyan")

		fmt.Printf("-> %s %s\n", m, u)
		err := c.Next()
		fmt.Printf("<- %s %s\n", m, u)
		fmt.Printf("   %s\n", FmtColor(time.Since(start).String(), "yellow"))

		return err
	}

}

func SecondMiddleware() HandlerFunc {
	return func(c *Context) error {
		fmt.Println("second middleware")
		fmt.Println("stop chain")
		return c.Text("stop chain")
		// return next(c)
	}
}

func withLogging(c *Context) error {
	start := time.Now()
	m := FmtColor(c.Req.Method, "green")
	u := FmtColor(c.Req.URL.String(), "cyan")

	fmt.Printf("-> %s %s\n", m, u)
	err := c.Next()
	fmt.Printf("<- %s %s\n", m, u)
	fmt.Printf("   %s\n", FmtColor(time.Since(start).String(), "yellow"))

	return err
}
