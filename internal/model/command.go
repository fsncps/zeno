package model

import "time"

// Command represents a single stored command/snippet joined with its language.
type Command struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	CodeMD      string    `db:"code_md"`
	Keyword     string    `db:"keyword"`
	Count       int       `db:"count"`
	UpdatedOn   time.Time `db:"updated_on"`
	Language    string    `db:"language"` // COALESCE(l.lang_name,'') AS language
}
