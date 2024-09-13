package middleware

import (
	"fmt"
	"time"

	"github.com/jocades/lit"
	"github.com/jocades/lit/color"
)

func Logging() lit.HandlerFunc {
	return func(c *lit.Context) error {
		start := time.Now()
		m := color.Magenta(c.Req.Method)
		u := color.Cyan(c.Req.URL.String())

		fmt.Printf("-> %s %s\n", m, u)
		err := c.Next()

		fmt.Println("after next", err, c.Res)
		status := c.Res.Status
		elapsed := time.Since(start).String()

		fmt.Printf("<- %s %s\n", m, u)
		fmt.Printf("   %s %s\n", lit.FmtStatus(status), lit.FmtColor(elapsed, "blue"))

		return err
	}

}
