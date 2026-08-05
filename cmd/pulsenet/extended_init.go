package main

import (
	"fmt"
	"os"

	"github.com/appoloncel283-debug/pulsenet/internal/ui"
)

func init() {
	if version == "2.2.0" || version == "2.2.1" {
		version = "2.3.0"
	}
	if len(os.Args) < 2 {
		return
	}

	var err error
	handled := true
	switch os.Args[1] {
	case "db":
		err = databaseCommand(os.Args[2:])
	case "site-diff":
		err = siteDiffCommand(os.Args[2:])
	case "site-secrets":
		err = siteSecretsCommand(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Println(`Additional PulseNet 2.3 commands:
  pulsenet db tools
  pulsenet db schema --engine <engine> --database <name-or-file> --output <file>
  pulsenet db backup --engine <engine> --database <name-or-file> --output <file>
  pulsenet db verify --engine <engine> --file <backup>
  pulsenet site-diff --old <dump-directory> --new <dump-directory>
  pulsenet site-secrets --dump <dump-directory>
`)
		handled = false
	default:
		handled = false
	}

	if !handled {
		return
	}
	if err != nil {
		ui.PrintError(err)
		os.Exit(1)
	}
	os.Exit(0)
}
