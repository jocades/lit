package lit_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jocades/lit"
)

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

	t.Run("returns text", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/text", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertBody(t, response.Body.String(), "Hello World!")
	})

	t.Run("returns json", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/json", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		// assertBody(t, response.Body.String(), `{"message":"Hello World!"}`)
	})

	t.Run("returns html", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/html", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertBody(t, response.Body.String(), "<h1>Hello World!</h1>")
	})

	t.Run("returns bad request", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/bad", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns internal error", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/err", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})

}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct body, got %s, want %s", got, want)
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if got == nil {
		t.Fatal("wanted an error but didn't get one")
	}

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
