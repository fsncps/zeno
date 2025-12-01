package main

import (
	"fmt"
	"os"

	"github.com/fsncps/zeno/internal/tui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// default: search UI
		if err := tui.RunSearch(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	switch args[0] {
	case "search":
		if err := tui.RunSearch(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "add":
		if err := tui.RunAdd(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Zeno is a command and snippet cheat sheet manager. Write your
	pet peeve commands to a DB with sexy lipgloss on them and retrieve when needed.

Usage:
  zeno            Start search UI
  zeno search     Start search UI (explicit)
  zeno add        Add a new command snippet
  zeno help       Show this help

Environment:
  ZENO_NO_AI=1    Disable AI helpers when adding commands
`)
}
