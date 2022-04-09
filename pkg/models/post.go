package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hoenn/whohirin/pkg/data"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Foreground(lipgloss.Color("#ff6600"))
	}()

	readMarkerStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Foreground(lipgloss.Color("#00FF00"))
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
)

type PostModel struct {
	postID            string
	previous          tea.Model
	commentIDs        []string
	currentComment    *data.Comment
	currentCommentIdx int

	hn        *data.Fetcher
	formatter formatter

	help         help.Model
	keys         postKeyMap
	viewport     viewport.Model
	windowHeight int
	ready        bool
}

func NewPost(postID string, previous tea.Model, fetcher *data.Fetcher) (*PostModel, error) {
	commentIDs, err := fetcher.PostCommentsList(postID)
	if err != nil {
		return nil, fmt.Errorf("unable to load post")
	}
	h := help.New()
	h.Width = initWindowSize.Width
	m := &PostModel{
		postID:            postID,
		hn:                fetcher,
		previous:          previous,
		commentIDs:        commentIDs,
		currentCommentIdx: 0,
		help:              h,
		keys:              postKeys,
		ready:             false,
		formatter: formatter{
			Width: defaultWidth,
		},
	}

	// Initial Viewport setup
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	helpHeight := lipgloss.Height(m.help.View(m.keys))
	verticalMarginHeight := headerHeight + footerHeight + helpHeight
	m.windowHeight = initWindowSize.Height
	m.viewport = viewport.New(initWindowSize.Width, m.windowHeight-verticalMarginHeight)
	m.formatter.Width = initWindowSize.Width - 1
	m.viewport.YPosition = headerHeight
	m.loadCurrentSelection()
	m.updateViews()

	m.ready = true

	return m, nil
}

type postKeyMap struct {
	Left   key.Binding
	Right  key.Binding
	Back   key.Binding
	Help   key.Binding
	Quit   key.Binding
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}

func (k postKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Back, k.Quit}
}
func (k postKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down, k.Select},
		{k.Help, k.Back, k.Quit},
	}
}

var postKeys = postKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("← / h", "previous"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→ / l", "next"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "previous view"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter / space", "mark as read"),
	),
	Help: key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h / ?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q / ctrl+c", "quit"),
	),
	// These are defined in the viewport component without help text.
	Up: key.NewBinding(
		key.WithHelp("↑ / k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithHelp("↓ / j", "scroll down"),
	),
}

func (m PostModel) Init() tea.Cmd {
	return nil
}

func (m PostModel) View() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView(), m.help.View(m.keys))
}

func (m PostModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.decCommentSelection()
			m.updateViews()
		case key.Matches(msg, m.keys.Right):
			m.incCommentSelection()
			m.updateViews()
		case key.Matches(msg, m.keys.Back):
			return m.previous, nil
		case key.Matches(msg, m.keys.Select):
			m.loadCurrentSelection()
			m.toggleCurrentSelectionRead()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.updateViews()
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		helpHeight := lipgloss.Height(m.help.View(m.keys))
		verticalMarginHeight := headerHeight + footerHeight + helpHeight
		m.help.Width = msg.Width

		if !m.ready {
			m.windowHeight = msg.Height
			m.viewport = viewport.New(msg.Width, m.windowHeight-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.formatter.Width = msg.Width - 1
			m.updateViews()
			m.viewport.YPosition = headerHeight + 1
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.windowHeight = msg.Height
			m.viewport.Height = m.windowHeight - verticalMarginHeight
			m.formatter.Width = msg.Width - 1
			m.updateViews()
		}

	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *PostModel) updateViews() *PostModel {
	if m.currentComment == nil {
		return m
	}

	text := m.currentComment.Data.Text
	if m.currentComment.Data.Text == "" {
		text = fmt.Sprintf("%d was [deleted]", m.currentComment.Data.ID)
	}

	m.viewport.YOffset = 0
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	helpHeight := lipgloss.Height(m.help.View(m.keys))
	verticalMarginHeight := headerHeight + footerHeight + helpHeight
	m.viewport.Height = m.windowHeight - verticalMarginHeight
	m.viewport.SetContent(m.formatter.Text(text))
	return m
}

func (m *PostModel) headerView() string {
	author := "?"
	if m.currentComment != nil {
		author = m.currentComment.Data.By
	}
	title := titleStyle.Render(fmt.Sprintf("https://news.ycombinator.com/user?id=%s", author))

	readStatus := "◯"
	if m.currentComment != nil {
		if m.currentComment.Read == true {
			readStatus = "✔"
		}
	}
	readMarker := readMarkerStyle.Render(readStatus)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)-lipgloss.Width(readMarker)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, readMarker)
}

func (m *PostModel) footerView() string {
	var info string
	if m.viewport.AtBottom() {
		info = infoStyle.Render("✔")
	} else {
		info = infoStyle.Render("🡳")
	}
	currentNumberInfo := infoStyle.Render(fmt.Sprintf("%d / %d", m.currentCommentIdx+1, len(m.commentIDs)))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)-lipgloss.Width(currentNumberInfo)))
	return lipgloss.JoinHorizontal(lipgloss.Center, currentNumberInfo, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *PostModel) incCommentSelection() {
	m.currentCommentIdx = (m.currentCommentIdx + 1) % len(m.commentIDs)
	m.loadCurrentSelection()
}

func (m *PostModel) decCommentSelection() {
	if m.currentCommentIdx-1 < 0 {
		m.currentCommentIdx = len(m.commentIDs) - 1
	}
	m.currentCommentIdx = m.currentCommentIdx - 1
	m.loadCurrentSelection()
}

func (m *PostModel) loadCurrentSelection() {
	c, err := m.hn.PostComment(m.postID, m.commentIDs[m.currentCommentIdx])
	if err != nil {
		fmt.Println(err)
		return
	}
	m.currentComment = c
}

func (m *PostModel) toggleCurrentSelectionRead() {
	c := m.hn.TogglePostCommentRead(m.postID, m.commentIDs[m.currentCommentIdx])
	m.currentComment = c
}
