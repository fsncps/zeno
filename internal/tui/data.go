package tui

import (
	"context"
	"database/sql"
	"strings"

	"github.com/fsncps/zeno/internal/db"
)

// Shared DTO for UI.
type commandItem struct {
	id         int
	title      string
	desc       string
	code       string
	keywords   string
	count      int
	lastUsed   string
	language   string // DISPLAY (caps)
	formatters string // comma-separated
}

func (i commandItem) Title() string       { return i.title }
func (i commandItem) Description() string { return i.desc }
func (i commandItem) FilterValue() string { return i.title + " " + i.desc + " " + i.keywords }

// FetchCommands reads all commands with language metadata.
func FetchCommands(dbh *sql.DB) ([]commandItem, error) {
	rows, err := dbh.Query(`
    SELECT c.ID, c.title, c.description, c.code_md, c.keywords, c.count, c.updated_on,
           UPPER(COALESCE(l.lang_name, '')) AS lang_name,
           COALESCE(l.formatter_bin, '')     AS formatter_bin
      FROM command c
 LEFT JOIN language l ON l.id = c.lang_id
  ORDER BY c.count DESC, c.updated_on DESC`)
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
			id:         id,
			title:      title,
			desc:       desc,
			code:       code,
			keywords:   keywords,
			count:      count,
			lastUsed:   updatedOn,
			language:   strings.ToUpper(langName),
			formatters: formatterBins,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCommandByID deletes a command row by ID with a short-lived connection.
func DeleteCommandByID(id int) error {
	ctx := context.Background()
	conn, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "DELETE FROM command WHERE id = ?", id)
	return err
}

func UpdateCommand(id int, title, desc, kw, code string) error {
    ctx := context.Background()
    conn, err := db.Connect(ctx)
    if err != nil {
        return err
    }
    defer conn.Close()

    _, err = conn.ExecContext(ctx, `
        UPDATE command
        SET title=?, description=?, keywords=?, code=?, updated_on=NOW()
        WHERE id=?`,
        title, desc, kw, code, id,
    )
    return err
}


