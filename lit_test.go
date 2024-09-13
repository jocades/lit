package lit_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jocades/lit"
	"github.com/stretchr/testify/assert"
)

func makeRequest(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestLit(t *testing.T) {
	app := lit.New()

	app.GET("/text", func(c *lit.Context) error {
		return c.Text("Hello World!")
	})

	app.GET("/json", func(c *lit.Context) error {
		return c.JSON(lit.Map{"message": "Hello World!"})
	})

	app.GET("/html", func(c *lit.Context) error {
		return c.HTML("<h1>Hello World!</h1>")
	})

	app.GET("/bad", func(c *lit.Context) error {
		return lit.ErrBadRequest
	})

	app.GET("/err", func(c *lit.Context) error {
		return errors.New("internal error")
	})

	// make use of the htttest and assert packages to test endpoints
	t.Run("returns text", func(t *testing.T) {
		assert.HTTPSuccess(t, app.ServeHTTP, "GET", "/text", nil)
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/text", nil, "Hello World!")
	})

	t.Run("returns json", func(t *testing.T) {
		assert.HTTPSuccess(t, app.ServeHTTP, "GET", "/json", nil)
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/json", nil, "Hello World!")
	})

	t.Run("returns html", func(t *testing.T) {
		assert.HTTPSuccess(t, app.ServeHTTP, "GET", "/html", nil)
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/html", nil, "Hello World!")
	})

	t.Run("returns bad request", func(t *testing.T) {
		res := makeRequest(app, "GET", "/bad")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("returns internal error", func(t *testing.T) {
		res := makeRequest(app, "GET", "/err")
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

}

func TestLitMW(t *testing.T) {
	app := lit.New()
	app.Use(lit.RequestLogger())

	app.GET("/",
		func(c *lit.Context) error {
			t.Log("mw1")
			return c.Next()
		},
		func(c *lit.Context) error {
			t.Log("mw2")
			return c.Text("Hello World!")
		})

	app.GET("/next",
		func(c *lit.Context) error {
			c.Next() // should error here
			return c.Next()
		},
		func(c *lit.Context) error {
			return c.Text("Hello World!")
		})

	t.Run("middleware", func(t *testing.T) {
		assert.HTTPSuccess(t, app.ServeHTTP, "GET", "/", nil)
	})

	t.Run("multiple next calls", func(t *testing.T) {
		res := makeRequest(app, "GET", "/next")
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}
