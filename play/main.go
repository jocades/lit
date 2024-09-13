package main

import (
	"fmt"
	"github.com/jocades/lit"
	"net/http"
)

func main() {
	app := lit.New()

	app.Use(lit.RequestLogger())
	// app.Use(lit.SecondMiddleware())

	app.GET("/",
		func(c *lit.Context) error {
			fmt.Println("mw1")
			// c.Next()
			return c.Next()
		},
		func(c *lit.Context) error {
			fmt.Println("mw2")
			// c.Next()
			return c.Text("Hello World!")
		})

	fmt.Println("Listening...")

	if err := http.ListenAndServe(":8000", app); err != nil {
		panic(err)
	}

}
