package main

import (
	"fmt"
	"os"

	"github.com/fsncps/zeno/internal/config"
	"github.com/fsncps/zeno/internal/tui"
	"github.com/fsncps/zeno/internal/version"
)

func main() {
	showConfig := false
	args := os.Args[1:]

	for i, arg := range args {
		if arg == "--show-config" {
			showConfig = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) > 0 && args[0] == "--version" {
		fmt.Println(version.Version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if showConfig {
		fmt.Printf("Config file: %s\n", cfg.ConfigFile)
		fmt.Printf("Env file: %s\n", cfg.EnvFile)
		fmt.Printf("Database: %s@%s:%s/%s\n", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
		fmt.Printf("Theme: %s\n", cfg.Theme)
		if len(cfg.OpenAIKey) > 10 {
			fmt.Printf("OpenAI: %s***\n", cfg.OpenAIKey[:10])
		} else if cfg.OpenAIKey != "" {
			fmt.Printf("OpenAI: ***\n")
		} else {
			fmt.Printf("OpenAI: (not set)\n")
		}
		return
	}

	if len(args) == 0 {
		if err := tui.RunSearch(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	switch args[0] {
	case "search":
		if err := tui.RunSearch(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "add":
		if err := tui.RunAdd(cfg); err != nil {
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
	fmt.Print(`Zeno is a command and snippet cheat sheet manager. Write your
oft-used commands to a DB and retrieve them when needed.

Usage:
  zeno            Start search UI
  zeno search     Start search UI (explicit)
  zeno add        Add a new command snippet
  zeno help       Show this help

Environment:
  ZENO_NO_AI=1    Disable AI helpers when adding commands

Config:
  ~/.config/zeno/config.yaml    General settings
  ~/.config/zeno/.env           Secrets (DB credentials, API keys)

Flags:
  --show-config    Print current configuration and exit
`)
}
