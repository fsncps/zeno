package main

import (
    "fmt"
    "os"

    "github.com/fsncps/zeno/internal/tui"
)

func main() {
    args := os.Args[1:]
    if len(args) > 0 && args[0] == "search" {
        if err := tui.RunSearch(); err != nil {
            fmt.Fprintln(os.Stderr, "Error:", err)
            os.Exit(1)
        }
        return
    }
    // default to add
    if err := tui.RunAdd(); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

