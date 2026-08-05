package main

import (
	"flag"
	"fmt"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
	"github.com/appoloncel283-debug/pulsenet/internal/ui"
)

func routerCommand(args []string) error {
	subcommand := "info"
	if len(args) > 0 && args[0] != "--json" && args[0] != "-h" && args[0] != "--help" {
		subcommand = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("router "+subcommand, flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	switch subcommand {
	case "info", "detect":
		return ui.RunRouterAssistant(false, *jsonOutput)
	case "open":
		if *jsonOutput {
			return fmt.Errorf("--json cannot be combined with router open")
		}
		return ui.RunRouterAssistant(true, false)
	default:
		return fmt.Errorf("usage: pulsenet router <info|open> [--json]")
	}
}

func integrityCommand(args []string, status core.IntegrityStatus) error {
	fs := flag.NewFlagSet("integrity", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ui.PrintIntegrityStatus(status, *jsonOutput)
	if status.State == "mismatch" || status.State == "error" {
		return fmt.Errorf("integrity state is %s", status.State)
	}
	return nil
}
