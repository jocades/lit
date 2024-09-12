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

type App struct {
	mux          *http.ServeMux
	ErrorHandler ErrHandlerFunc
}

func New() *App {
	return &App{
		mux:          http.NewServeMux(),
		ErrorHandler: handleError,
	}
}

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

func addRoute(mux *http.ServeMux, method, path string, h HandlerFunc, eh ErrHandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}

	mux.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		// the best place to create the conext should be in the ServeHTTP method
		// of the App struct but since we are using the http.ServeMux
		// for routing I cannot find a way to do that.
		// leave it here for now but this dissallows the possibility of creating
		// 'subapps' or 'subrouters' that can be later 'merged' like in
		// the case of the http.ServeMux type since it only has to commply with the
		// http.Handler interface. so we sould keep all the logic in the ServeHTTP method
		// of the App struct. but for now, this is fine.
		c := &Context{
			Req:   r,
			Res:   NewResponse(w),
			Query: r.URL.Query(),
		}

		if err := h(c); err != nil {
			eh(err, c)
		}
	})
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m := FmtColor(r.Method, "green")
		// p := FmtColor(r.URL.Path, "cyan")
		p := FmtColor(r.URL.String(), "cyan")

		fmt.Printf("-> %s %s\n", m, p)

		h.ServeHTTP(w, r)

		fmt.Printf("<- %s %s\n", m, p)
		fmt.Printf("   %s\n", FmtColor(time.Since(start).String(), "yellow"))
	})
}

func (a *App) GET(path string, h HandlerFunc) {
	addRoute(a.mux, http.MethodGet, path, h, a.ErrorHandler)
}

func (a *App) POST(path string, h HandlerFunc) {
	addRoute(a.mux, http.MethodPost, path, h, a.ErrorHandler)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	WithLogging(a.mux).ServeHTTP(w, r)
}

// Response implements the ResponseWriter interface, it is a thin wrapper
// to have more control over the response.
//
// A ResponseWriter interface is used by an HTTP handler to
// construct an HTTP response.
type Response struct {
	writer   http.ResponseWriter
	Status   int
	Commited bool
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{writer: w}
}

// Header returns the header map that will be sent by
// [ResponseWriter.WriteHeader]. The [Header] map also is the mechanism with which
// [Handler] implementations can set HTTP trailers.
//
// Changing the header map after a call to [ResponseWriter.WriteHeader] (or
// [ResponseWriter.Write]) has no effect unless the HTTP status code was of the
// 1xx class or the modified headers are trailers.
//
// There are two ways to set Trailers. The preferred way is to
// predeclare in the headers which trailers you will later
// send by setting the "Trailer" header to the names of the
// trailer keys which will come later. In this case, those
// keys of the Header map are treated as if they were
// trailers. See the example. The second way, for trailer
// keys not known to the [Handler] until after the first [ResponseWriter.Write],
// is to prefix the [Header] map keys with the [TrailerPrefix]
// constant value.
//
// To suppress automatic response headers (such as "Date"), set
// their value to nil.
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
//
// If [ResponseWriter.WriteHeader] has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data. If the Header
// does not contain a Content-Type line, Write adds a Content-Type set
// to the result of passing the initial 512 bytes of written data to
// [DetectContentType]. Additionally, if the total size of all written
// data is under a few KB and there are no Flush calls, the
// Content-Length header is added automatically.
//
// Depending on the HTTP protocol version and the client, calling
// Write or WriteHeader may prevent future reads on the
// Request.Body. For HTTP/1.x requests, handlers should read any
// needed request body data before writing the response. Once the
// headers have been flushed (due to either an explicit Flusher.Flush
// call or writing enough data to trigger a flush), the request body
// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
// handlers to continue to read the request body while concurrently
// writing the response. However, such behavior may not be supported
// by all HTTP/2 clients. Handlers should read before writing if
// possible to maximize compatibility.
func (r *Response) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
//
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes or 1xx informational responses.
//
// The provided code must be a valid HTTP 1xx-5xx status code.
// Any number of 1xx headers may be written, followed by at most
// one 2xx-5xx header. 1xx headers are sent immediately, but 2xx-5xx
// headers may be buffered. Use the Flusher interface to send
// buffered data. The header map is cleared when 2xx-5xx headers are
// sent, but not with 1xx headers.
//
// The server will automatically send a 100 (Continue) header
// on the first read from the request body if the request has
// an "Expect: 100-continue" header.
func (r *Response) WriteHeader(code int) {
	r.writer.WriteHeader(code)
}

type Context struct {
	Req   *http.Request
	Res   *Response
	Query url.Values
}

func (c *Context) writeHeader(status []int) {
	// the status code is automatically set to 200 if not set
	if len(status) > 0 {
		c.Res.WriteHeader(status[0])
	}
}

func (c *Context) Body(b []byte, ct string, status ...int) error {
	c.Res.Header().Set(HeaderContentType, ct)
	c.writeHeader(status)

	_, err := c.Res.Write(b)
	return err
}

func (c *Context) Text(s string, status ...int) error {
	return c.Body([]byte(s), MIMETextPlain, status...)
}

func (c *Context) HTML(s string, status ...int) error {
	return c.Body([]byte(s), MIMETextHTML, status...)
}

func (c *Context) JSON(v any, status ...int) error {
	c.Res.Header().Set(HeaderContentType, MIMEApplicationJSON)
	c.writeHeader(status)

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
