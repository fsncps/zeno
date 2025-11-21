package tui

import (
	"context"
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsncps/zeno/internal/db"
)

// Fetch rows from DB with language join.
func fetchCommands(conn *sql.DB) ([]commandItem, error) {
	rows, err := conn.Query(`
		SELECT
			c.id,
			c.title,
			c.description,
			c.code_md,
			c.keywords,
			c.count,
			c.updated_on,
			COALESCE(l.lang_name, '')     AS lang_name,
			COALESCE(l.formatter_bin, '') AS formatter_bin
		FROM command c
		LEFT JOIN language l ON l.id = c.lang_id
		ORDER BY c.count DESC, c.updated_on DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []commandItem
	for rows.Next() {
		var (
			id, count               int
			title, desc, code       string
			keywords, updatedOn     string
			langName, formatterBins string
		)
		if err := rows.Scan(&id, &title, &desc, &code, &keywords, &count, &updatedOn, &langName, &formatterBins); err != nil {
			return nil, err
		}
		out = append(out, commandItem{
			id:         id, // <<< THIS WAS MISSING
			title:      title,
			desc:       desc,
			code:       code,
			keywords:   keywords,
			count:      count,
			lastUsed:   updatedOn,
			language:   langName,
			formatters: formatterBins,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Entry point: connect, load, render.
func RunSearch() error {
	conn, err := db.Connect(context.Background())
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}
	defer conn.Close()

	cmds, err := fetchCommands(conn)
	if err != nil {
		return err
	}
	if len(cmds) == 0 {
		fmt.Println("No commands in database.")
		return nil
	}

	p := tea.NewProgram(newSearchModel(cmds, 0, 0))
	return p.Start()
}
