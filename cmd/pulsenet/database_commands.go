package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/ui"
)

type repeatedStringFlag []string

func (values *repeatedStringFlag) String() string { return fmt.Sprint([]string(*values)) }
func (values *repeatedStringFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func databaseCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pulsenet db <tools|schema|backup|verify> [options]")
	}
	switch args[0] {
	case "tools":
		fs := flag.NewFlagSet("db tools", flag.ContinueOnError)
		jsonOutput := fs.Bool("json", false, "print JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		ui.RunDBTools(*jsonOutput)
		return nil
	case "schema", "backup":
		operation := args[0]
		fs := flag.NewFlagSet("db "+operation, flag.ContinueOnError)
		engine := fs.String("engine", "", "postgres, mysql, or sqlite")
		database := fs.String("database", "", "database name, connection string, or SQLite file")
		output := fs.String("output", "", "output file")
		timeout := fs.Duration("timeout", 30*time.Minute, "operation timeout")
		jsonOutput := fs.Bool("json", false, "print JSON")
		var extraArgs repeatedStringFlag
		fs.Var(&extraArgs, "arg", "additional official client argument; may be repeated")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *engine == "" || *database == "" || *output == "" {
			return fmt.Errorf("usage: pulsenet db %s --engine <engine> --database <name-or-file> --output <file>", operation)
		}
		return ui.RunDBExport(*engine, operation, *database, *output, extraArgs, *timeout, *jsonOutput)
	case "verify":
		fs := flag.NewFlagSet("db verify", flag.ContinueOnError)
		engine := fs.String("engine", "", "postgres, mysql, or sqlite")
		path := fs.String("file", "", "backup or dump file")
		timeout := fs.Duration("timeout", 2*time.Minute, "verification timeout")
		jsonOutput := fs.Bool("json", false, "print JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *engine == "" || *path == "" {
			return fmt.Errorf("usage: pulsenet db verify --engine <engine> --file <path>")
		}
		return ui.RunDBVerify(*engine, *path, *timeout, *jsonOutput)
	default:
		return fmt.Errorf("unknown db subcommand %q", args[0])
	}
}

func siteDiffCommand(args []string) error {
	fs := flag.NewFlagSet("site-diff", flag.ContinueOnError)
	oldDirectory := fs.String("old", "", "older site-dump directory")
	newDirectory := fs.String("new", "", "newer site-dump directory")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *oldDirectory == "" || *newDirectory == "" {
		return fmt.Errorf("usage: pulsenet site-diff --old <directory> --new <directory>")
	}
	return ui.RunSiteDiff(*oldDirectory, *newDirectory, *jsonOutput)
}

func siteSecretsCommand(args []string) error {
	fs := flag.NewFlagSet("site-secrets", flag.ContinueOnError)
	directory := fs.String("dump", "", "local site-dump directory")
	maxMB := fs.Int64("max-mb", 32, "maximum size per scanned file in MiB")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *directory == "" {
		return fmt.Errorf("usage: pulsenet site-secrets --dump <directory>")
	}
	if *maxMB < 1 || *maxMB > 256 {
		return fmt.Errorf("max-mb must be between 1 and 256")
	}
	return ui.RunSiteSecretScan(*directory, *maxMB, *jsonOutput)
}
