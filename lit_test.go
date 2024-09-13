package lit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jocades/lit"
	"github.com/jocades/lit/middleware"
	"github.com/stretchr/testify/assert"
)

func makeRequest(h http.Handler, m, p string, b any, withAuth bool) *httptest.ResponseRecorder {
	body, _ := json.Marshal(b)
	req, _ := http.NewRequest(m, p, bytes.NewBuffer(body))
	if withAuth {
		req.Header.Set("Authorization", "Bearer token")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w

}

func decode[T any](r io.Reader) (T, error) {
	var v T
	return v, json.NewDecoder(r).Decode(&v)
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
		w := makeRequest(app, "GET", "/text", nil, false)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("returns json", func(t *testing.T) {
		w := makeRequest(app, "GET", "/json", nil, false)
		body, _ := decode[lit.Map](w.Body)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Hello World!", body["message"])
	})

	t.Run("returns html", func(t *testing.T) {
		w := makeRequest(app, "GET", "/html", nil, false)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "<h1>Hello World!</h1>", w.Body.String())
	})

	t.Run("returns bad request", func(t *testing.T) {
		res := makeRequest(app, "GET", "/bad", nil, false)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("returns internal error", func(t *testing.T) {
		res := makeRequest(app, "GET", "/err", nil, false)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

}

func TestLitMW(t *testing.T) {
	app := lit.New()
	app.Use(middleware.Logging())

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
		w := makeRequest(app, "GET", "/", nil, false)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	// t.Run("multiple next calls", func(t *testing.T) {
	// 	res := makeRequest(app, "GET", "/next")
	// 	assert.Equal(t, http.StatusInternalServerError, res.Code)
	// })
}
