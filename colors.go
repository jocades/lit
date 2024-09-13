package lit

import "fmt"

func FmtColor(v any, color string) string {
	switch color {
	case "red":
		return fmt.Sprintf("\033[31m%v\033[0m", v)
	case "green":
		return fmt.Sprintf("\033[32m%v\033[0m", v)
	case "yellow":
		return fmt.Sprintf("\033[33m%v\033[0m", v)
	case "blue":
		return fmt.Sprintf("\033[34m%v\033[0m", v)
	case "magenta":
		return fmt.Sprintf("\033[35m%v\033[0m", v)
	case "cyan":
		return fmt.Sprintf("\033[36m%v\033[0m", v)
	case "white":
		return fmt.Sprintf("\033[37m%v\033[0m", v)
	case "gray":
		return fmt.Sprintf("\033[90m%v\033[0m", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
