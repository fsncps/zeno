package tui

import (
	"encoding/json"
	"fmt"

	"github.com/fsncps/zeno/internal/strutil"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/fsncps/zeno/internal/ai"
	"github.com/fsncps/zeno/internal/db"
)

func RunAdd() error {
	var title string
	var mode string
	var code string

	// 1. Title prompt
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Command title").
				Value(&title),
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

	// 2. Handle clipboard vs manual entry
	if mode == "clip" {
		clip, err := clipboard.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read clipboard: %w", err)
		}
		code = clip
	} else {
		// Manual: run another form with multiline text
		form2 := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Title("Code snippet").
					Lines(10).
					Value(&code),
			),
		)
		if err := form2.Run(); err != nil {
			return err
		}
	}

	// AI: Generate title, description, keywords
	// Always format the title locally first
	formattedTitle := strutil.TitleCase(title)

	// AI: Generate description and keywords
	title, desc, kws, err := ai.SummarizeAndKeywords(title, code)
	if err != nil {
		fmt.Println("AI error, using fallback:", err)
		desc = "(todo: description)"
		kws = []string{}
	}

	kwsJSON, _ := json.Marshal(kws)
	embedding := "[]"

	conn := db.Connect()
	defer conn.Close()

	_, err = conn.Exec(`
          INSERT INTO command (title, description, code_md, keywords, embedding)
          VALUES (?, ?, ?, ?, ?)
      `, formattedTitle, desc, code, string(kwsJSON), embedding)

	if err != nil {
		return err
	}

	fmt.Println("Command added successfully.")
	return nil
}
