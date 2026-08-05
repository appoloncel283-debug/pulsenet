package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunMenu(version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		clear()
		printLogo(version)
		fmt.Println(Color(bold, "  1") + "  Full diagnosis")
		fmt.Println(Color(bold, "  2") + "  Availability monitor")
		fmt.Println(Color(bold, "  3") + "  DNS records and resolver comparison")
		fmt.Println(Color(bold, "  4") + "  TLS certificate inspection")
		fmt.Println(Color(bold, "  5") + "  Security header audit")
		fmt.Println(Color(bold, "  6") + "  Explicit port check")
		fmt.Println(Color(bold, "  7") + "  HTTP benchmark")
		fmt.Println(Color(bold, "  8") + "  Site dump")
		fmt.Println(Color(bold, "  9") + "  Website log viewer")
		fmt.Println(Color(bold, " 10") + "  Support the project")
		fmt.Println(Color(bold, "  0") + "  Exit")
		fmt.Print("\nSelect an action: ")
		switch readLine(reader) {
		case "1":
			target := ask(reader, "Website, domain, or IP: ")
			if target != "" {
				_ = RunDiagnose(target, 5*time.Second, 3, nil, false, "", "")
			}
			pause(reader)
		case "2":
			target := ask(reader, "Website URL: ")
			seconds, _ := strconv.Atoi(askDefault(reader, "Interval in seconds", "5"))
			if seconds < 1 {
				seconds = 5
			}
			RunWatch(target, time.Duration(seconds)*time.Second, 5*time.Second, 0, "", false)
			pause(reader)
		case "3":
			RunDNS(ask(reader, "Domain or IP: "), 5*time.Second, false)
			pause(reader)
		case "4":
			RunTLS(ask(reader, "Domain or host:port: "), 5*time.Second, false)
			pause(reader)
		case "5":
			RunHeaders(ask(reader, "Website URL: "), 8*time.Second, false, false)
			pause(reader)
		case "6":
			host := ask(reader, "Host: ")
			ports := askDefault(reader, "Ports", "22,80,443,3389")
			RunPorts(host, ports, 2*time.Second, 32, false, false)
			pause(reader)
		case "7":
			url := ask(reader, "Website URL: ")
			RunBenchmark(url, "GET", 20, 4, 10*time.Second, false, false)
			pause(reader)
		case "8":
			target := ask(reader, "Website URL: ")
			if target != "" {
				if err := RunSiteDump(target, "", 16*1024*1024, 15*time.Second, false, false); err != nil {
					PrintError(err)
				}
			}
			pause(reader)
		case "9":
			path := ask(reader, "Local log file: ")
			if path != "" {
				if err := RunLogs(path, 100, false, 500*time.Millisecond, core.LogFilter{}, false); err != nil {
					PrintError(err)
				}
			}
			pause(reader)
		case "10":
			PrintSupport()
			pause(reader)
		case "0", "q", "quit", "exit":
			return
		default:
			fmt.Println(Color(red, "Unknown selection."))
			time.Sleep(700 * time.Millisecond)
		}
	}
}
