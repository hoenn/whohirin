package models

import (
	"fmt"
	"strings"

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

	viewport viewport.Model
	ready    bool
}

const defaultWidth = 80

func NewPost(postID string, previous tea.Model, fetcher *data.Fetcher) (*PostModel, error) {
	commentIDs, err := fetcher.PostCommentsList(postID)
	if err != nil {
		return nil, fmt.Errorf("unable to load post")
	}

	s := &PostModel{
		postID:            postID,
		hn:                fetcher,
		previous:          previous,
		commentIDs:        commentIDs,
		currentCommentIdx: 0,
		ready:             false,
		formatter: formatter{
			Width: defaultWidth,
		},
	}

	// Initial Viewport setup
	headerHeight := lipgloss.Height(s.headerView())
	footerHeight := lipgloss.Height(s.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	s.viewport = viewport.New(initWindowSize.Width, initWindowSize.Height-verticalMarginHeight)
	s.formatter.Width = initWindowSize.Width - 1
	s.viewport.YPosition = headerHeight
	s.loadCurrentSelection()
	s.updateViews()

	s.ready = true

	return s, nil
}

func (m PostModel) Init() tea.Cmd {
	return nil
}

func (m PostModel) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m PostModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.decCommentSelection()
			m.updateViews()
		case "right", "l":
			m.incCommentSelection()
			m.updateViews()
		case "backspace":
			return m.previous, nil
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.formatter.Width = msg.Width - 1
			m.updateViews()
			m.viewport.YPosition = headerHeight + 1
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
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
	m.viewport.SetContent(m.formatter.Text(text))
	return m
}

func (m *PostModel) headerView() string {
	author := "?"
	if m.currentComment != nil {
		author = m.currentComment.Data.By
	}
	title := titleStyle.Render(fmt.Sprintf("https://news.ycombinator.com/user?id=%s", author))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
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
