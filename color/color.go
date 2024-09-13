package color

import "fmt"

func Red(v any) string {
	return fmt.Sprintf("\033[31m%v\033[0m", v)
}

func Green(v any) string {
	return fmt.Sprintf("\033[32m%v\033[0m", v)
}

func Yellow(v any) string {
	return fmt.Sprintf("\033[33m%v\033[0m", v)
}

func Blue(v any) string {
	return fmt.Sprintf("\033[34m%v\033[0m", v)
}

func Magenta(v any) string {
	return fmt.Sprintf("\033[35m%v\033[0m", v)
}

func Cyan(v any) string {
	return fmt.Sprintf("\033[36m%v\033[0m", v)
}
