package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsncps/zeno/internal/db"
)

const maxPreviewLines = 20

// --- Data model ---

type commandItem struct {
	title, desc, code string
	keywords          string
	count             int
	lastUsed          string
}

func (i commandItem) Title() string       { return i.title }
func (i commandItem) Description() string { return i.desc }
func (i commandItem) FilterValue() string { return i.title + " " + i.desc }

// --- Bubble Tea model ---

type searchModel struct {
	list    list.Model
	details string
	width   int
	height  int
}

// highlightCode: syntax highlighting with Chroma, using Catppuccin Macchiato.
func highlightCode(code string, lang string) string {
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	formatter := formatters.TTY16m
	style := styles.Get("catppuccin-macchiato")
	if style == nil {
		style = styles.Fallback
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return code
	}
	return buf.String()
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
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return searchModel{
		list:    l,
		details: cmds[0].code,
		width:   width,
		height:  height,
	}
}

func (m searchModel) Init() tea.Cmd { return nil }

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		leftWidth := int(0.4 * float32(msg.Width))
		if leftWidth < 20 {
			leftWidth = msg.Width / 2
		}
		m.list.SetSize(leftWidth-2, msg.Height-2)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			sel := m.list.SelectedItem().(commandItem)
			_ = clipboard.WriteAll(sel.code)
			fmt.Println("Copied to clipboard")
			return m, tea.Quit
		}
	}

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

	paddingLeft := "  "
	border := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	footerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(rightWidth)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	sel, _ := m.list.SelectedItem().(commandItem)

	hr := paddingLeft + strings.Repeat("─", rightWidth-2)

	header := fmt.Sprintf(
		"%s\n\n%s%s\n\n%s%s\n\n%s",
		hr,
		paddingLeft, titleStyle.Render(sel.title),
		paddingLeft, sel.desc,
		hr,
	)

	// Code block (truncated to 20 lines for view only)
	codeLines := strings.Split(sel.code, "\n")
	if len(codeLines) > maxPreviewLines {
		codeLines = codeLines[:maxPreviewLines]
		codeLines = append(codeLines, "... (truncated)")
	}
	for i, l := range codeLines {
		if len(l) > rightWidth-2 {
			codeLines[i] = l[:rightWidth-2]
		}
	}
	highlighted := highlightCode(strings.Join(codeLines, "\n"), "")
	codeBlock := highlighted

	meta := fmt.Sprintf(
		"Keywords: %s\nHit count: %d\nLast used: %s",
		sel.keywords,
		sel.count,
		sel.lastUsed,
	)
	footer := footerBox.Render(meta)

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

// --- Entry point ---

func RunSearch() error {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query(`
    SELECT ID, title, description, code_md, keywords, count, updated_on
    FROM command
    ORDER BY count DESC, updated_on DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cmds []commandItem
	for rows.Next() {
		var id, count int
		var title, desc, code, keywords, updatedOn string
		if err := rows.Scan(&id, &title, &desc, &code, &keywords, &count, &updatedOn); err != nil {
			return err
		}
		cmds = append(cmds, commandItem{
			title:    title,
			desc:     desc,
			code:     code,
			keywords: keywords,
			count:    count,
			lastUsed: updatedOn,
		})
	}

	if len(cmds) == 0 {
		fmt.Println("No commands in database.")
		return nil
	}

	p := tea.NewProgram(
		newSearchModel(cmds, 0, 0),
	)
	return p.Start()
}
