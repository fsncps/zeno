package main

import (
    "fmt"
    "os"

    "github.com/fsncps/zeno/internal/tui"
)

func main() {
    if err := tui.RunAdd(); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

