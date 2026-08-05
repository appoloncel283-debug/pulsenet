package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func PrintSupport() {
	fmt.Println("\n" + Color(yellow+bold, "      ◉  SUPPORT PULSENET"))
	fmt.Println(Color(dim, "  USDT on the TRON network (TRC20)"))
	fmt.Println()
	fmt.Println("  " + Color(bold, "TGVDhCbDKEnWV5BVrUtMicjhwMiJVUYSSh"))
	fmt.Println()
	fmt.Println(Color(dim, "  Always verify the network is TRC20 before sending."))
}

func PrintJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		PrintError(err)
	}
}

func PrintError(err error) { fmt.Fprintln(os.Stderr, Color(red, "Error: "+err.Error())) }

func printRecord(name string, values []string) {
	if len(values) == 0 {
		return
	}
	for i, value := range values {
		label := ""
		if i == 0 {
			label = name
		}
		fmt.Printf("  %-7s %s\n", label, value)
	}
}

func printLogo(version string) {
	fmt.Println(Color(cyan+bold, "  PulseNet"), Color(dim, "v"+version+" — practical network diagnostics"))
	fmt.Println(Color(dim, "  ─────────────────────────────────────────────────"))
	fmt.Println()
}

func clear() { fmt.Print("\033[2J\033[H") }
func readLine(reader *bufio.Reader) string {
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}
func ask(reader *bufio.Reader, label string) string { fmt.Print(label); return readLine(reader) }
func askDefault(reader *bufio.Reader, label, fallback string) string {
	fmt.Printf("%s [%s]: ", label, fallback)
	value := readLine(reader)
	if value == "" {
		return fallback
	}
	return value
}
func pause(reader *bufio.Reader) {
	fmt.Print("\nPress Enter to return to the menu...")
	_, _ = reader.ReadString('\n')
}
func compact(value string, max int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max-1]) + "…"
}
func emptyAs(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
