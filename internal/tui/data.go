package tui

import (
	"context"
	"database/sql"
	"strings"
)

type commandItem struct {
	id         int
	title      string
	desc       string
	code       string
	keywords   string
	count      int
	lastUsed   string
	language   string
	formatters string
}

func (i commandItem) Title() string       { return i.title }
func (i commandItem) Description() string { return i.desc }
func (i commandItem) FilterValue() string { return i.title + " " + i.desc + " " + i.keywords }

func DeleteCommandByID(conn *sql.DB, id int) error {
	_, err := conn.ExecContext(context.Background(), "DELETE FROM command WHERE id = ?", id)
	return err
}

func RecordSearchUsage(term string, commandID int, conn *sql.DB) error {
	term = strings.TrimSpace(term)

	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

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

	if term == "" {
		return tx.Commit()
	}

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

	return tx.Commit()
}
