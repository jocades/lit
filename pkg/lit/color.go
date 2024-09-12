package lit

import "fmt"

func FmtColor(s string, color string) string {
	switch color {
	case "red":
		return fmt.Sprintf("\033[31m%s\033[0m", s)
	case "green":
		return fmt.Sprintf("\033[32m%s\033[0m", s)
	case "yellow":
		return fmt.Sprintf("\033[33m%s\033[0m", s)
	case "blue":
		return fmt.Sprintf("\033[34m%s\033[0m", s)
	case "magenta":
		return fmt.Sprintf("\033[35m%s\033[0m", s)
	case "cyan":
		return fmt.Sprintf("\033[36m%s\033[0m", s)
	default:
		return s
	}
}
