package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fsncps/zeno/internal/db"
)

type modifyModel struct {
	width  int
	height int

	id          int
	title       *textarea.Model
	desc        *textarea.Model
	keywords    *textarea.Model
	code        *textarea.Model

	field  int
	status string // optional status line (errors)
}

const (
	fieldTitle = iota
	fieldDesc
	fieldKw
	fieldCode
)

// styles
var (
	sectionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	footerStyle = lipgloss.NewStyle().
			MarginTop(0).
			Foreground(lipgloss.Color("240"))

	statusErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)

func newTA(h int) *textarea.Model {
	ta := textarea.New()
	ta.SetHeight(h)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	return &ta
}

func newModifyModel(ci commandItem, w, h int) modifyModel {
	t := newTA(1)
	t.SetValue(ci.title)

	d := newTA(3)
	d.SetValue(ci.desc)

	k := newTA(3)
	k.SetValue(ci.keywords)

	c := newTA(4) // will be resized by geometry
	c.SetValue(ci.code)

	m := modifyModel{
		width:    w,
		height:   h,
		id:       ci.id,
		title:    t,
		desc:     d,
		keywords: k,
		code:     c,
		field:    fieldTitle,
	}
	// apply geometry immediately if we already know width/height
	if m.width > 0 && m.height > 0 {
		m.applyGeometry()
	}
	return m.focus(fieldTitle)
}

func (m modifyModel) focus(n int) modifyModel {
	m.title.Blur()
	m.desc.Blur()
	m.keywords.Blur()
	m.code.Blur()

	m.field = n
	switch n {
	case fieldTitle:
		m.title.Focus()
	case fieldDesc:
		m.desc.Focus()
	case fieldKw:
		m.keywords.Focus()
	case fieldCode:
		m.code.Focus()
	}
	return m
}

func (m modifyModel) Init() tea.Cmd { return nil }

func (m modifyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.applyGeometry()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEsc:
			// discard changes, go back to search
			model, cmd := m.backToSearch()
			return model, cmd

		case tea.KeyCtrlS:
			// save then go back to search
			if err := m.saveSync(); err != nil {
				m2 := m
				m2.status = "Save failed: " + err.Error()
				return m2, nil
			}
			model, cmd := m.backToSearch()
			return model, cmd

		case tea.KeyTab:
			return m.focus((m.field+1)%4), nil

		case tea.KeyShiftTab:
			return m.focus((m.field+3)%4), nil
		}
	}

	// delegate to focused textarea
	switch m.field {
	case fieldTitle:
		var cmd tea.Cmd
		*m.title, cmd = m.title.Update(msg)
		return m, cmd
	case fieldDesc:
		var cmd tea.Cmd
		*m.desc, cmd = m.desc.Update(msg)
		return m, cmd
	case fieldKw:
		var cmd tea.Cmd
		*m.keywords, cmd = m.keywords.Update(msg)
		return m, cmd
	case fieldCode:
		var cmd tea.Cmd
		*m.code, cmd = m.code.Update(msg)
		return m, cmd
	}

	return m, nil
}

// applyGeometry: compute widths & heights from screen size
func (m *modifyModel) applyGeometry() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	// width: use almost full screen
	innerW := m.width - 4
	if innerW < 20 {
		innerW = m.width
	}
	m.title.SetWidth(innerW)
	m.desc.SetWidth(innerW)
	m.keywords.SetWidth(innerW)
	m.code.SetWidth(innerW)

	// height budgeting (textarea heights):
	// title: 1, desc: 3, kw: 3, code: remainder (min 3)
	// height budgeting (textarea heights):
	// title: 1, desc: 3, kw: 3, code: remainder (min 3)
	titleH := 1
	descH := 3
	kwH := 3

	// more conservative allowance for box borders + labels + footer/status.
	// Estimated:
	//   4 boxes * (1 label + 2 borders) = 12
	//   footer + possible status          ~ 2–3
	// So use 16 to stay below the screen height.
	overhead := 16

	codeH := m.height - (titleH + descH + kwH + overhead)
	if codeH < 3 {
		codeH = 3
	}

	m.title.SetHeight(titleH)
	m.desc.SetHeight(descH)
	m.keywords.SetHeight(kwH)
	m.code.SetHeight(codeH)

}

func (m modifyModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "EDIT\n" + m.title.View()
	}

	innerW := m.width - 4
	if innerW < 20 {
		innerW = m.width
	}

	titleSection := sectionBoxStyle.Width(innerW).Render(
		labelStyle.Render("TITLE") + "\n" + m.title.View(),
	)

	descSection := sectionBoxStyle.Width(innerW).Render(
		labelStyle.Render("DESCRIPTION") + "\n" + m.desc.View(),
	)

	kwSection := sectionBoxStyle.Width(innerW).Render(
		labelStyle.Render("KEYWORDS") + "\n" + m.keywords.View(),
	)

	codeSection := sectionBoxStyle.Width(innerW).Render(
		labelStyle.Render("CODE") + "\n" + m.code.View(),
	)

	status := ""
	if m.status != "" {
		status = statusErrStyle.Render(m.status)
	}

	footer := footerStyle.Render(
		"<Ctrl+S> Save    •    <Esc> Cancel    •    <Tab/Shift+Tab> Move between fields",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleSection,
		descSection,
		kwSection,
		codeSection,
		status,
		footer,
	)
}

// saveSync: update DB synchronously
func (m modifyModel) saveSync() error {
	id := m.id
	title := strings.TrimSpace(m.title.Value())
	desc := strings.TrimSpace(m.desc.Value())
	kw := strings.TrimSpace(m.keywords.Value())
	code := m.code.Value()

	ctx := context.Background()
	conn, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `
        UPDATE command
        SET title = ?, description = ?, keywords = ?, code_md = ?, updated_on = NOW()
        WHERE id = ?`,
		title, desc, kw, code, id,
	)
	return err
}

// backToSearch: reload commands and return a fresh searchModel
func (m modifyModel) backToSearch() (tea.Model, tea.Cmd) {
	ctx := context.Background()
	conn, err := db.Connect(ctx)
	if err != nil {
		m2 := m
		m2.status = "DB reconnect failed: " + err.Error()
		return m2, nil
	}
	defer conn.Close()

	cmds, err := fetchCommands(conn)
	if err != nil || len(cmds) == 0 {
		m2 := m
		if err != nil {
			m2.status = "Reload failed: " + err.Error()
		} else {
			m2.status = "No commands found."
		}
		return m2, nil
	}

	return newSearchModel(cmds, m.width, m.height), nil
}

