# zeno

A tiny but super simple and fast command and code snippet cheat sheet manager to build a custom database with the oft-used commands, becuase it feels better than asking ChatGPT every time. Designed to be lean and as efficient as possible. Made with Go and the [Charm/huh/lipgloss](https://github.com/charmbracelet/lipgloss) libs.

## Usage
```
$ zeno help
Zeno is a command and snippet cheat sheet manager. Write your
pet peeve commands to a DB with sexy lipgloss on them and and retrieve when needed.

Usage:
  zeno            Start search UI
  zeno search     Start search UI (explicit)
  zeno add        Add a new command snippet
  zeno help       Show this help

Environment:
  ZENO_NO_AI=1    Disable AI helpers when adding commands
```
- `zeno [search]` opens a TUI screen: left side has a list of commands with a smart, fuzzy filter active to type, right side shows the command with highlighting and formatting plus meta-info. Selecting a command with Enter copies it to clipboard and returns to the shell.
- `zeno add` loads a small input form, where you enter a short title for your command plus the command itself (command is taken from clipboard if already copied) and the language. The languages have smart sorting and the right one should be on top. OpenAI completion endpoints will then phrase a short description of the command add some keywords, and the record is saved to MySQL/MariaDB.

## Deps & Installation

You need Go and MariaDB and have to have the DB set up. Create a DB and user, then import the schema from the repo you clone with `git clone https://github.com/fsncps/zeno`.
Then, to install just cd into the repo root and run `make` and `make install`.

---
