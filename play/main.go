package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jocades/lit"
	mw "github.com/jocades/lit/middleware"
)

func main() {
	app := lit.New()

	app.Use(mw.Logging())
	app.Use(func(c *lit.Context) error {
		fmt.Println("use-mw")
		return c.Next()
	})

	app.GET("/",
		func(c *lit.Context) error {
			fmt.Println("mw1")
			c.Next()
			return c.Next()
		},
		func(c *lit.Context) error {
			fmt.Println("mw2")
			return c.Text("Hello World!")
		})

	app.GET("/not", func(c *lit.Context) error {
		return c.NotFound()
	})

	fmt.Println("Listening...")

	mux := http.NewServeMux()
	sub := http.NewServeMux()

	// h := func(res http.ResponseWriter, r *http.Request) {
	// 	res.Write([]byte("Hello World!"))
	// }

	// hf := http.HandlerFunc(h)

	sub.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "root")
	})

	sub.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "sub")
	})

	mux.Handle("/", sub)

	if err := http.ListenAndServe(":8000", app); err != nil {
		log.Fatal(err)
	}

}
