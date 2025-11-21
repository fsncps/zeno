package chroma

import (
	"bytes"

	chroma2 "github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightCode returns ANSI-colored code for terminal output.
// If highlighting fails, it falls back to the original input.
func HighlightCode(code, language string) string {
	if code == "" {
		return ""
	}

	var lexer chroma2.Lexer

	if language != "" {
		lexer = lexers.Get(language)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("catppuccin-mocha")
	if style == nil {
		style = styles.Fallback
	}

	// NOTE: TTY is already a Formatter value, not a function.
	formatter := formatters.TTY

	it, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, it); err != nil {
		return code
	}

	return buf.String()
}

// Highlight is a thin alias kept for backwards compatibility.
func Highlight(code, language string) string {
	return HighlightCode(code, language)
}
