package lit

import (
	// "errors"
	// "log"
	// "net/http"
	//
	_ "github.com/jocades/lit/pkg"
)

/* func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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

	log.Println("Listening...")

	if err := http.ListenAndServe(":8000", app); err != nil {
		log.Fatal(err)
	}
} */
