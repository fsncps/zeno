package tui

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"regexp"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/fsncps/zeno/internal/ai"
	"github.com/fsncps/zeno/internal/db"
)

type Language struct {
	ID    int
	Name  string
	Desc  string
	Count int
}


func fetchLanguages(ctx context.Context, conn *sql.DB) ([]Language, error) {
	const q = `
SELECT l.id, l.lang_name, l.lang_desc, COALESCE(COUNT(c.id),0) AS cmd_count
FROM language l
LEFT JOIN command c ON c.lang_id = l.id
GROUP BY l.id, l.lang_name, l.lang_desc
`
	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var langs []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.ID, &l.Name, &l.Desc, &l.Count); err != nil {
			return nil, err
		}
		langs = append(langs, l)
	}
	return langs, rows.Err()
}


func RunAdd() error {
	ctx := context.Background()
	conn, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var title, mode, code string

	// 1) Title + input mode
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Command title").Value(&title),
			huh.NewSelect[string]().
				Title("Code input").
				Options(
					huh.NewOption("Clipboard", "clip"),
					huh.NewOption("Manual entry", "manual"),
				).
				Value(&mode),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	// 2) Code
	if mode == "clip" {
		clip, err := clipboard.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read clipboard: %w", err)
		}
		code = clip
	} else {
		if err := huh.NewForm(huh.NewGroup(
			huh.NewText().Title("Code snippet").Lines(10).Value(&code),
		)).Run(); err != nil {
			return err
		}
	}

	// 3) Language (from DB, ordered by substring relevance incl. desc)
	dbLangs, err := fetchLanguages(ctx, conn)
	if err != nil { return err }
	if len(dbLangs) == 0 { return fmt.Errorf("no languages found in DB") }

	ordered := orderLangsByTitleSubstring(dbLangs, title)

	var selectedLang string
	opts := make([]huh.Option[string], 0, len(ordered))
	for _, l := range ordered {
		label := fmt.Sprintf("%-10s  %-45s  [%3d]",
			strings.ToUpper(l.Name),
			l.Desc,
			l.Count,
		)
		opts = append(opts, huh.NewOption(label, l.Name))
	}




	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Language").
			Options(opts...).
			Value(&selectedLang),
	)).Run(); err != nil {
		return err
	}

	// map to ID
	var langID int
	for _, l := range dbLangs {
		if l.Name == selectedLang {
			langID = l.ID
			break
		}
	}
	if langID == 0 {
		return fmt.Errorf("selected language not found: %q", selectedLang)
	}


	// 4) AI summary
	desc, kws, aiErr := ai.SummarizeAndKeywords(title, code)
	if aiErr != nil {
		fmt.Println("AI error, using fallback:", aiErr)
		desc = "(todo: description)"
		kws = []string{}
	}
	// prepend lang to keywords if missing
	if selectedLang != "" {
		found := false
		for _, k := range kws {
			if strings.EqualFold(k, selectedLang) {
				found = true
				break
			}
		}
		if !found {
			kws = append([]string{selectedLang}, kws...)
		}
	}

	kwsJSON, _ := json.Marshal(kws)
	embedding := "[]"

	// 5) Insert
	_, err = conn.Exec(`
		INSERT INTO command (title, description, code_md, keywords, embedding, lang_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, title, desc, code, string(kwsJSON), embedding, langID)
	if err != nil {
		return err
	}

	fmt.Println("Command added successfully.")
	return nil
}

// ---------- helpers ----------

func matchScore(l Language, title string) int {
	t := strings.ToLower(title)
	name := strings.ToLower(l.Name)

	// precompile simple word-boundary regex for the lang name
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)

	score := 0

	// 1) Name match: whole word >> substring; ignore tiny substrings (len<3)
	if re.FindStringIndex(t) != nil {
		score += len(name) * 100 // dominant weight
	} else if len(name) >= 3 && strings.Contains(t, name) {
		score += len(name) * 10
	} // else: ignore short incidental substrings like "c" in "code" or "go" in "go to"

	// 2) Desc words: only words >=3 chars; whole-word only; take longest hit
	// split desc into words
	desc := strings.ToLower(l.Desc)
	longest := 0
	start := -1
	for i, r := range desc {
		isWord := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isWord && start == -1 {
			start = i
		}
		if (!isWord || i == len(desc)-1) && start != -1 {
			end := i
			if isWord && i == len(desc)-1 {
				end = i + 1
			}
			w := desc[start:end]
			start = -1
			if len(w) >= 3 {
				// whole-word match in title
				wre := regexp.MustCompile(`\b` + regexp.QuoteMeta(w) + `\b`)
				if wre.FindStringIndex(t) != nil && len(w) > longest {
					longest = len(w)
				}
			}
		}
	}
	if longest > 0 {
		score += longest // small tie-breaker vs name
	}

	return score
}

func orderLangsByTitleSubstring(langs []Language, title string) []Language {
	type wrap struct {
		L     Language
		Score int
	}
	ws := make([]wrap, 0, len(langs))
	for _, l := range langs {
		ws = append(ws, wrap{L: l, Score: matchScore(l, title)})
	}

	// partition into matches vs others
	var matches, others []wrap
	for _, w := range ws {
		if w.Score > 0 {
			matches = append(matches, w)
		} else {
			others = append(others, w)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score // stronger/longer match first
		}
		if matches[i].L.Count != matches[j].L.Count {
			return matches[i].L.Count > matches[j].L.Count
		}
		return matches[i].L.Name < matches[j].L.Name
	})
	sort.Slice(others, func(i, j int) bool {
		if others[i].L.Count != others[j].L.Count {
			return others[i].L.Count > others[j].L.Count
		}
		return others[i].L.Name < others[j].L.Name
	})

	out := make([]Language, 0, len(langs))
	for _, w := range matches {
		out = append(out, w.L)
	}
	for _, w := range others {
		out = append(out, w.L)
	}
	return out
}
