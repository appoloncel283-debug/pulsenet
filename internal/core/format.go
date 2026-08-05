package core

import "fmt"

func FormatMS(value float64) string {
	if value < 1 {
		return fmt.Sprintf("%.2f ms", value)
	}
	return fmt.Sprintf("%.1f ms", value)
}
