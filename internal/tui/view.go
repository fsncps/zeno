package tui

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxPreviewLines = 20
	rightPadLeft    = 2 // header & description left padding
)

// ---------------- UI model ----------------

type searchModel struct {
	list            list.Model
	allItems        []list.Item // source items for filtering
	details         string
	query           string
	width           int
	height          int
	confirmDelete   bool
	pendingDeleteID int
	confirmMsg      string
}

func newSearchModel(cmds []commandItem, width, height int) searchModel {
	items := make([]list.Item, len(cmds))
	for i, c := range cmds {
		items[i] = c
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true

	l := list.New(items, delegate, 0, 0)
	l.Title = "Commands"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // we do our own filtering
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	initial := ""
	if len(cmds) > 0 {
		initial = cmds[0].code
	}

	return searchModel{
		list:     l,
		allItems: items,
		details:  initial,
		width:    width,
		height:   height,
	}
}

// strict substring filter: all query tokens must appear in FilterValue() + code
func (m *searchModel) applyFilter() {
	q := strings.TrimSpace(m.query)
	if q == "" {
		m.list.SetItems(m.allItems)
		return
	}

	tokens := strings.Fields(strings.ToLower(q))
	if len(tokens) == 0 {
		m.list.SetItems(m.allItems)
		return
	}

	out := make([]list.Item, 0, len(m.allItems))
	for _, it := range m.allItems {
		ci, ok := it.(commandItem)
		if !ok {
			continue
		}
		// include code in search base (title + desc + keywords + code)
		fv := strings.ToLower(ci.FilterValue() + "\n" + ci.code)

		match := true
		for _, tok := range tokens {
			if !strings.Contains(fv, tok) {
				match = false
				break
			}
		}
		if match {
			out = append(out, it)
		}
	}
	m.list.SetItems(out)
}

// remove item from allItems and re-apply current filter
func (m *searchModel) removeItemByID(id int) {
	filtered := make([]list.Item, 0, len(m.allItems))
	for _, it := range m.allItems {
		ci, ok := it.(commandItem)
		if !ok {
			continue
		}
		if ci.id != id {
			filtered = append(filtered, it)
		}
	}
	m.allItems = filtered
	m.applyFilter()
}

func (m searchModel) Init() tea.Cmd { return nil }

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ------------------------------------------------------------
	// WINDOW SIZE
	// ------------------------------------------------------------
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		leftWidth := int(0.4 * float32(msg.Width))
		if leftWidth < 20 {
			leftWidth = msg.Width / 2
		}
		m.list.SetSize(leftWidth-2, msg.Height-2)
		return m, nil

	// ------------------------------------------------------------
	// KEY HANDLING
	// ------------------------------------------------------------
	case tea.KeyMsg:
		// 1) control / special keys by Type
		switch msg.Type {

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			// cancel delete confirmation first, if active
			if m.confirmDelete {
				m2 := m
				m2.confirmDelete = false
				m2.pendingDeleteID = 0
				m2.confirmMsg = ""
				return m2, nil
			}
			// then clear query if non-empty
			if m.query != "" {
				m2 := m
				m2.query = ""
				m2.applyFilter()
				return m2, nil
			}
			// otherwise quit
			return m, tea.Quit

		case tea.KeyCtrlD:
			// start delete confirmation on current selection
			if !m.confirmDelete {
				if sel, ok := m.list.SelectedItem().(commandItem); ok {
					m2 := m
					m2.confirmDelete = true
					m2.pendingDeleteID = sel.id
					m2.confirmMsg = fmt.Sprintf(
						"Delete command %q (id=%d)? Press Y to confirm, N or ESC to cancel.",
						sel.title, sel.id,
					)
					return m2, nil
				}
			}

		case tea.KeyBackspace:
			// don't modify filter while confirm dialog is up
			if !m.confirmDelete && m.query != "" {
				m2 := m
				_, n := utf8.DecodeLastRuneInString(m2.query)
				if n > 0 && n <= len(m2.query) {
					m2.query = m2.query[:len(m2.query)-n]
				}
				m2.applyFilter()
				return m2, nil
			}

		case tea.KeyEnter:
			// only copy when not in delete confirmation mode
			if sel, ok := m.list.SelectedItem().(commandItem); ok && !m.confirmDelete {
				_ = clipboard.WriteAll(sel.code)
				fmt.Println("Copied to clipboard")
				return m, tea.Quit
			}
		}

		// 2) delete-confirm character handling (y/Y, n/N)
		if m.confirmDelete {
			s := msg.String()
			switch s {
			case "y", "Y", "z", "Z": // keep z/Z if you want to be safe for QWERTZ
				if m.pendingDeleteID != 0 {
					id := m.pendingDeleteID

					if err := DeleteCommandByID(id); err != nil {
						m2 := m
						m2.confirmDelete = false
						m2.pendingDeleteID = 0
						m2.confirmMsg = "Delete failed: " + err.Error()
						return m2, nil
					}

					m2 := m
					m2.removeItemByID(id)
					m2.confirmDelete = false
					m2.pendingDeleteID = 0
					m2.confirmMsg = ""
					return m2, nil
				}

			case "n", "N":
				m2 := m
				m2.confirmDelete = false
				m2.pendingDeleteID = 0
				m2.confirmMsg = ""
				return m2, nil
			}

			// any other key while confirming:
			// - do NOT feed to filter (see below, gated by !m.confirmDelete)
			// - DO fall through to list.Update so arrows etc. still work
		} else {

			// 3) type-to-search: ONLY when not in confirm mode and printable input
			if len(msg.Runes) == 1 &&
				unicode.IsPrint(msg.Runes[0]) &&
				msg.Runes[0] != '\n' &&
				msg.Runes[0] != '\t' {

				m2 := m
				m2.query += string(msg.Runes[0])
				m2.applyFilter()
				return m2, nil
			}
		}
	}

	// ------------------------------------------------------------
	// DEFAULT FALLTHROUGH → let list handle leftover messages
	// ------------------------------------------------------------
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if sel, ok := m.list.SelectedItem().(commandItem); ok {
		m.details = sel.code
	}
	return m, cmd
}

func (m searchModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return m.list.View()
	}

	leftWidth := int(0.4 * float32(m.width))
	rightWidth := m.width - leftWidth - 2
	totalHeight := m.height - 2

	border := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	footerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(rightWidth)

	// titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	matchStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

	wrapBox := lipgloss.NewStyle().Width(rightWidth - 2).PaddingLeft(rightPadLeft)

	codePadLeft := rightPadLeft / 2
	if codePadLeft < 1 {
		codePadLeft = 1
	}
	codeBox := lipgloss.NewStyle().Width(rightWidth - 2).PaddingLeft(codePadLeft)

	// Selected item (safe if list empty)
	var sel commandItem
	if it := m.list.SelectedItem(); it != nil {
		if ci, ok := it.(commandItem); ok {
			sel = ci
		}
	}

	// tokens for highlighting (only title/desc/meta, not code)
	var tokens []string
	q := strings.TrimSpace(m.query)
	if q != "" {
		tokens = strings.Fields(strings.ToLower(q))
	}

	hr := strings.Repeat("─", rightWidth-2)

	titleText := highlightTokens(sel.title, tokens, matchStyle)
	descText := highlightTokens(sel.desc, tokens, matchStyle)

	header := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		hr,
		wrapBox.Render(titleText),
		wrapBox.Render(descText),
		hr,
	)

	// preview code block (truncated for preview)
	codeLines := strings.Split(sel.code, "\n")
	if len(codeLines) > maxPreviewLines {
		codeLines = codeLines[:maxPreviewLines]
		codeLines = append(codeLines, "... (truncated)")
	}
	highlighted := highlightCode(strings.Join(codeLines, "\n"), strings.ToLower(sel.language))
	codeBlock := codeBox.Render(highlighted)

	langDisp := strings.ToUpper(sel.language)
	if langDisp == "" {
		langDisp = "-"
	}
	formatters := sel.formatters
	if formatters == "" {
		formatters = "-"
	}

	metaKeywords := highlightTokens(sel.keywords, tokens, matchStyle)

	info := fmt.Sprintf(
		"Language:      %s\nFormatters:    %s\nKeywords:      %s\nHit count:     %d\nLast used:     %s",
		langDisp, formatters, metaKeywords, sel.count, sel.lastUsed)

	if m.confirmDelete && m.confirmMsg != "" {
		info += "\n" + m.confirmMsg
	} else {
		info += "\n\nKEYS:          <↑/↓> to select  •  <CR> to copy & return  •  <ESC> to clear/quit  •  <Ctrl+D> to delete\n\nFilter is active!\n"
	}
	info += fmt.Sprintf(
		"Query: %q",
		m.query,
	)

	footer := footerBox.Render(info)

	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	codeHeight := lipgloss.Height(codeBlock)
	freeSpace := totalHeight - headerHeight - footerHeight - codeHeight
	if freeSpace < 0 {
		freeSpace = 0
	}
	topPad := freeSpace / 2

	rightView := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		strings.Repeat("\n", topPad),
		codeBlock,
		strings.Repeat("\n", freeSpace-topPad),
		footer,
	)

	listView := border.Width(leftWidth).Render(m.list.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, rightView)
}

// highlightTokens wraps any occurrence of tokens in s with the given style.
// case-insensitive, but preserves original casing.
func highlightTokens(s string, tokens []string, style lipgloss.Style) string {
	if s == "" || len(tokens) == 0 {
		return s
	}
	out := s
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		lowerOut := strings.ToLower(out)
		var sb strings.Builder
		i := 0
		for {
			idx := strings.Index(lowerOut[i:], tok)
			if idx == -1 {
				sb.WriteString(out[i:])
				break
			}
			idx += i
			sb.WriteString(out[i:idx])
			end := idx + len(tok)
			if end > len(out) {
				end = len(out)
			}
			sb.WriteString(style.Render(out[idx:end]))
			i = end
		}
		out = sb.String()
	}
	return out
}

// ---------------- highlighting ----------------

func normalizeLangAlias(lang string) string {
	l := strings.TrimSpace(strings.ToLower(lang))
	switch l {
	case "ps", "pwsh", "powershell", "ps1":
		return "powershell"
	case "js", "node", "nodejs":
		return "javascript"
	case "ts":
		return "typescript"
	case "sh", "shell", "zsh", "bash":
		return "bash"
	case "py":
		return "python"
	case "c#":
		return "csharp"
	case "c++":
		return "cpp"
	case "md", "markdown":
		return "markdown"
	case "yml":
		return "yaml"
	case "tf", "terraform":
		return "hcl"
	default:
		return l
	}
}

func pickLexer(lang, code string) chroma.Lexer {
	if lang != "" {
		if lx := lexers.Get(normalizeLangAlias(lang)); lx != nil {
			return lx
		}
		if lx := lexers.Match(lang); lx != nil {
			return lx
		}
	}
	if lx := lexers.Analyse(code); lx != nil {
		return lx
	}
	return lexers.Fallback
}

func highlightCode(code, lang string) string {
	lx := pickLexer(lang, code)
	it, err := lx.Tokenise(nil, code)
	if err != nil {
		return code
	}
	style := styles.Get("catppuccin-macchiato")
	if style == nil {
		style = styles.Fallback
	}
	var buf bytes.Buffer
	if err := formatters.TTY16m.Format(&buf, style, it); err != nil {
		return code
	}
	return buf.String()
}
