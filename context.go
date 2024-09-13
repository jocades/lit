package lit

import (
	"net/http"
	"net/url"
)

type Context struct {
	app    *App          // app instance
	Req    *http.Request // raw request
	Res    *Response     // wrapped response
	Header http.Header   // response header
	Query  url.Values    // request query
	Next   func() error  // next handler
}

func (c *Context) writeHeader(code []int) {
	if len(code) > 0 {
		c.Res.WriteHeader(code[0])
	}
}

func (c *Context) Send(b []byte, ct string, code ...int) error {
	c.Header.Set(HeaderContentType, ct)
	c.writeHeader(code)

	_, err := c.Res.Write(b)
	return err
}

func (c *Context) SendStatus(code int) error {
	c.Res.WriteHeader(code)
	return nil
}

func (c *Context) Text(s string, code ...int) error {
	return c.Send([]byte(s), MIMETextPlain, code...)
}

func (c *Context) HTML(s string, code ...int) error {
	return c.Send([]byte(s), MIMETextHTML, code...)
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
	return c.app.NotFoundHandler(c)
}
