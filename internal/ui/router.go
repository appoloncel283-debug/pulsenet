package ui

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunRouterAssistant(openBrowser, jsonOutput bool) error {
	info, err := core.InspectRouter(context.Background(), 6*time.Second)
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(info)
		return nil
	}
	printRouterInfo(info)
	if openBrowser {
		if err := core.OpenBrowser(info.AdminURL); err != nil {
			return err
		}
		fmt.Println(Color(green, "\nRouter admin page opened in the default browser."))
	}
	return nil
}

func RunRouterMenu(reader *bufio.Reader) {
	info, err := core.InspectRouter(context.Background(), 6*time.Second)
	if err != nil {
		PrintError(err)
		return
	}
	printRouterInfo(info)
	answer := strings.ToLower(askDefault(reader, "Open this router page in the browser", "Y"))
	if answer == "y" || answer == "yes" {
		if err := core.OpenBrowser(info.AdminURL); err != nil {
			PrintError(err)
			return
		}
		fmt.Println(Color(green, "Router admin page opened."))
	}
}

func printRouterInfo(info core.RouterInfo) {
	fmt.Println("\n" + Color(bold, "Router assistant"))
	fmt.Printf("  Gateway IP:    %s\n", info.GatewayIP)
	fmt.Printf("  Admin page:    %s\n", info.AdminURL)
	fmt.Printf("  Reachable:     %v\n", info.Reachable)
	if info.SSID != "" {
		fmt.Printf("  Wi-Fi name:    %s\n", info.SSID)
	}
	if info.GatewayMAC != "" {
		fmt.Printf("  Gateway MAC:   %s\n", info.GatewayMAC)
	}
	if info.HTTPStatus != "" {
		fmt.Printf("  HTTP status:   %s\n", info.HTTPStatus)
	}
	if info.PageTitle != "" {
		fmt.Printf("  Page title:    %s\n", info.PageTitle)
	}
	if info.ServerHeader != "" {
		fmt.Printf("  Server:        %s\n", info.ServerHeader)
	}
	for _, note := range info.Notes {
		fmt.Printf("  %s %s\n", Color(yellow, "→"), note)
	}
	for _, probeError := range info.ProbeErrors {
		fmt.Printf("  %s %s\n", Color(dim, "probe"), compact(probeError, 120))
	}
}
