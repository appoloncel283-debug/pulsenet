package ui

import "os"

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	dim    = "\033[2m"
)

func Color(code, text string) string {
	if os.Getenv("NO_COLOR") != "" {
		return text
	}
	return code + text + reset
}
