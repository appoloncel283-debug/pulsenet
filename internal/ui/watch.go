package ui

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunWatch(input string, interval, timeout time.Duration, count int, csvPath string, insecure bool) {
	target, err := core.ParseTarget(input)
	if err != nil {
		PrintError(err)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var file *os.File
	var csvWriter *csv.Writer
	if csvPath != "" {
		opened, actualWriter, err := core.OpenWatchCSV(csvPath)
		if err != nil {
			PrintError(err)
			return
		}
		file = opened
		csvWriter = actualWriter
		defer file.Close()
		defer csvWriter.Flush()
	}
	fmt.Printf("\nMonitoring %s every %s. Press Ctrl+C to stop.\n\n", target.Input, interval)
	var total, failures, consecutive, maxConsecutive int
	for {
		total++
		result := core.ProbeHTTP(ctx, target.CandidateURLs, timeout, insecure)
		sample := core.WatchSample{Timestamp: time.Now(), StatusCode: result.StatusCode, LatencyMS: result.Timings.TotalMS, Error: result.Error}
		sample.Up = result.Error == "" && result.StatusCode < 500
		if sample.Up {
			consecutive = 0
			fmt.Printf("%s  %s  HTTP %-3d  %4d ms\n", sample.Timestamp.Format("15:04:05"), Color(green, "UP  "), sample.StatusCode, sample.LatencyMS)
		} else {
			failures++
			consecutive++
			if consecutive > maxConsecutive {
				maxConsecutive = consecutive
			}
			message := sample.Error
			if message == "" {
				message = fmt.Sprintf("HTTP %d", sample.StatusCode)
			}
			fmt.Printf("%s  %s  %s\n", sample.Timestamp.Format("15:04:05"), Color(red, "DOWN"), compact(message, 110))
		}
		if csvWriter != nil {
			if err := core.WriteWatchCSV(csvWriter, sample); err != nil {
				PrintError(err)
			}
		}
		if count > 0 && total >= count {
			break
		}
		select {
		case <-ctx.Done():
			goto done
		case <-time.After(interval):
		}
	}

done:
	uptime := 100.0
	if total > 0 {
		uptime = float64(total-failures) / float64(total) * 100
	}
	fmt.Printf("\nChecks: %d · failures: %d · uptime: %.2f%% · longest outage: %d checks\n", total, failures, uptime, maxConsecutive)
	if csvPath != "" {
		fmt.Println(Color(green, "CSV log: "+csvPath))
	}
}
