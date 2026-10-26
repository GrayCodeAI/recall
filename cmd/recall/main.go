package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GrayCodeAI/recall/memory"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "recall:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("use remember, search, list, show, or forget")
	}
	cmd := args[0]
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storePath := fs.String("store", defaultStore(), "memory data directory")
	scope := fs.String("scope", "", "project scope")
	kind := fs.String("kind", "decision", "memory kind")
	text := fs.String("text", "", "memory content")
	source := fs.String("source", "", "source type and reference")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	store := memory.Store{Dir: *storePath}
	switch cmd {
	case "remember":
		r, err := store.Remember(*scope, *kind, *text, []string{*source})
		if err != nil {
			return err
		}
		return printJSON(r)
	case "search":
		if fs.NArg() == 0 {
			return errors.New("search query is required")
		}
		r, err := store.Search(*scope, fs.Arg(0))
		if err != nil {
			return err
		}
		return printJSON(r)
	case "list":
		r, err := store.List(*scope)
		if err != nil {
			return err
		}
		return printJSON(r)
	case "show":
		if fs.NArg() != 1 {
			return errors.New("show requires a memory ID")
		}
		r, err := store.Get(fs.Arg(0))
		if err != nil {
			return err
		}
		return printJSON(r)
	case "forget":
		if fs.NArg() != 1 {
			return errors.New("forget requires a memory ID")
		}
		return store.Forget(fs.Arg(0))
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printJSON(v any) error { return json.NewEncoder(os.Stdout).Encode(v) }

func defaultStore() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".recall"
	}
	return filepath.Join(dir, "recall", "memories")
}
