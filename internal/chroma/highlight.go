import (
	"bytes"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func highlightWithChroma(code string) string {
	var buf bytes.Buffer
	lexer := lexers.Analyse(code)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, _ := lexer.Tokenise(nil, code)
	formatter := formatters.TTY8Color // ANSI for terminal
	style := styles.Get("monokai")
	formatter.Format(&buf, style, iterator)
	return buf.String()
}

