package ui

import (
	"bufio"
	"fmt"
	"time"
)

func RunDatabaseMenu(reader *bufio.Reader) {
	for {
		clear()
		fmt.Println(Color(cyan+bold, "  PulseNet Database Toolkit"))
		fmt.Println(Color(dim, "  Authorized backups, schema exports, and integrity checks"))
		fmt.Println(Color(dim, "  ─────────────────────────────────────────────────────"))
		fmt.Println()
		fmt.Println(Color(bold, "  1") + "  Check database client tools")
		fmt.Println(Color(bold, "  2") + "  Export schema only")
		fmt.Println(Color(bold, "  3") + "  Create database backup")
		fmt.Println(Color(bold, "  4") + "  Verify backup or dump")
		fmt.Println(Color(bold, "  0") + "  Back")
		fmt.Print("\nSelect an action: ")

		switch readLine(reader) {
		case "1":
			RunDBTools(false)
			pause(reader)
		case "2", "3":
			operation := "schema"
			if readLastSelection(reader) == "3" {
				operation = "backup"
			}
		case "4":
			engine := askDefault(reader, "Engine (postgres, mysql, sqlite)", "sqlite")
			path := ask(reader, "Backup or dump file: ")
			if path != "" {
				if err := RunDBVerify(engine, path, 2*time.Minute, false); err != nil {
					PrintError(err)
				}
			}
			pause(reader)
		case "0", "b", "back":
			return
		default:
			fmt.Println(Color(red, "Unknown selection."))
			time.Sleep(700 * time.Millisecond)
		}
	}
}

// readLastSelection is intentionally unused outside this file; the export path
// is handled by runDatabaseExportPrompt below. It exists only to keep the menu
// switch easy to read.
func readLastSelection(_ *bufio.Reader) string { return "" }

func runDatabaseExportPrompt(reader *bufio.Reader, operation string) {
	engine := askDefault(reader, "Engine (postgres, mysql, sqlite)", "sqlite")
	database := ask(reader, "Database name, connection string, or SQLite file: ")
	output := ask(reader, "Output file: ")
	if database == "" || output == "" {
		fmt.Println(Color(yellow, "Database and output file are required."))
		return
	}
	if err := RunDBExport(engine, operation, database, output, nil, 30*time.Minute, false); err != nil {
		PrintError(err)
	}
}
