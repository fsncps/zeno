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
- `zeno [search]` opens a TUI screen: the left side has a list of commands with a smart, fuzzy filter active to type, while the reading pane on the right shows the command with highlighting and formatting, its title and description, and meta-info. Selecting a command with Enter copies it to clipboard and returns to the shell.
- `zeno add` loads a small input form, where you enter a short title for your command plus the command itself (command is taken from clipboard if already copied) and the language. The languages have smart sorting and the right one should be at the top. OpenAI completion endpoints will then phrase a short description of the command and add some keywords, and the record is saved to MySQL/MariaDB.

The filter and sorting mechanism is adaptive and will present hits ordered by (sort of) smart metrics. To be added is vercor space embedding of the codeblock records for similarity searching.

## Deps & Installation

You need Go and MariaDB and have to have the DB set up. Create a DB and user, set the env vars `$ZENODB_NAME`, `$ZENODB_USER`, `$ZENODB_PASS`; for remote DBs also `$ZENODB_HOST` and then import the schema from the repo you clone with `git clone https://github.com/fsncps/zeno`.
Then, to install just cd into the repo root and run `make` and `make install`.
For the AI description support, you need an OpenAI API key under `$OPENAI_API_KEY`.

---

