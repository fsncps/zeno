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

// RecordSearchUsage increments counters in search_term and search_hit
// for the given query term and command ID.
// No-op if term is empty.
// RecordSearchUsage increments counters in search_term, search_hit, and command
// for the given query term and command ID.
// No-op if term is empty.
// RecordSearchUsage increments command.count for the given command ID,
// and, if term is non-empty, also updates search_term and search_hit.
func RecordSearchUsage(term string, commandID int) error {
	term = strings.TrimSpace(term)

	ctx := context.Background()
	conn, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		// safe even after Commit()
		_ = tx.Rollback()
	}()

	// 1) Always increment the command's count (primary signal).
	_, err = tx.ExecContext(ctx, `
        UPDATE command
           SET count      = count + 1,
               updated_on = NOW()
         WHERE id = ?`,
		commandID,
	)
	if err != nil {
		return err
	}

	// 2) If no search term, we are done.
	term = strings.TrimSpace(term)
	if term == "" {
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}

	// 3) Upsert into search_term using unique key on term.
	res, err := tx.ExecContext(ctx, `
        INSERT INTO search_term (term, count)
        VALUES (?, 1)
        ON DUPLICATE KEY UPDATE
            count      = count + 1,
            updated_on = NOW(),
            id         = LAST_INSERT_ID(id)`,
		term,
	)
	if err != nil {
		return err
	}

	termID64, err := res.LastInsertId()
	if err != nil {
		return err
	}
	termID := int(termID64)

	// 4) Update or insert search_hit for (term_id, command_id).
	updRes, err := tx.ExecContext(ctx, `
        UPDATE search_hit
           SET count      = count + 1,
               updated_on = NOW()
         WHERE term_id = ? AND command_id = ?`,
		termID, commandID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := updRes.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO search_hit (term_id, command_id, count)
            VALUES (?, ?, 1)`,
			termID, commandID,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

