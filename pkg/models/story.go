package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hoenn/go-hn/pkg/hnapi"
	"github.com/muesli/reflow/wordwrap"
	"jaytaylor.com/html2text"
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

type StoryModel struct {
	storyID          int
	hn               *hnapi.HNClient
	previous         tea.Model
	comments         []int
	commentContent   map[int]*hnapi.Comment // lazy loaded
	currentSelection int
	currentComment   *hnapi.Comment
	formatter        *formatter

	viewport viewport.Model
	ready    bool
}

const defaultWidth = 80

func NewStory(storyID int, previous tea.Model, client *hnapi.HNClient) *StoryModel {
	item, err := client.Item(fmt.Sprintf("%d", storyID))
	if err != nil {
		panic(err)
	}
	story, ok := item.(*hnapi.Story)
	if !ok {
		panic(errors.New("fix this"))
	}

	s := &StoryModel{
		storyID:          storyID,
		hn:               client,
		previous:         previous,
		comments:         story.Kids,
		commentContent:   make(map[int]*hnapi.Comment),
		currentSelection: 0,
		ready:            false,
		formatter: &formatter{
			Width: defaultWidth,
		},
	}

	// Initial Viewport setup
	headerHeight := lipgloss.Height(s.headerView("??"))
	footerHeight := lipgloss.Height(s.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	s.viewport = viewport.New(initWindowSize.Width, initWindowSize.Height-verticalMarginHeight)
	s.viewport.YPosition = headerHeight
	s.currentComment = s.loadCommentContent()
	s.viewport.SetContent(s.formatCurrentContent())
	s.ready = true

	s.viewport.YPosition = headerHeight + 1

	return s
}

func (m StoryModel) Init() tea.Cmd {
	return nil
}

func (m StoryModel) View() string {
	if m.currentComment != nil {
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(m.currentComment.By), m.viewport.View(), m.footerView())
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView("?"), m.viewport.View(), m.footerView())
}

func (m *StoryModel) loadCommentContent() *hnapi.Comment {
	commentID := m.comments[m.currentSelection]
	content, ok := m.commentContent[commentID]
	if content == nil || !ok {
		cc, err := m.getComment(commentID)
		if err != nil {
			return nil
		}
		m.commentContent[commentID] = cc
		content = cc
	}
	return content
}

func (m StoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.currentSelection = m.decCommentSelection()
			m.currentComment = m.loadCommentContent()
			m.viewport.SetContent(m.formatter.Text(m.currentComment.Text))
		case "right", "l":
			m.currentSelection = m.incCommentSelection()
			m.currentComment = m.loadCommentContent()
			m.viewport.SetContent(m.formatter.Text(m.currentComment.Text))
		case "backspace":
			return m.previous, nil
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView("??"))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.currentComment = m.loadCommentContent()
			m.viewport.SetContent(m.formatter.Text(m.currentComment.Text))
			m.ready = true

			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m StoryModel) formatCurrentContent() string {
	var text string
	if m.currentComment != nil {
		t, err := html2text.FromString(m.currentComment.Text, html2text.Options{TextOnly: true})
		if err != nil {
			return "unable to format content"
		}
		text = t
	}
	if text == "" {
		return fmt.Sprintf("Comment %d, was [deleted]", m.currentComment.ID)
	}
	return wordwrap.String(text, 80)
}

func (m StoryModel) headerView(text string) string {
	title := titleStyle.Render(fmt.Sprintf("https://news.ycombinator.com/user?id=%s", text))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m StoryModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	currentNumberInfo := infoStyle.Render(fmt.Sprintf("%d / %d", m.currentSelection+1, len(m.comments)))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)-lipgloss.Width(currentNumberInfo)))
	return lipgloss.JoinHorizontal(lipgloss.Center, currentNumberInfo, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *StoryModel) getComment(id int) (*hnapi.Comment, error) {
	item, err := c.hn.Item(fmt.Sprintf("%d", id))
	if err != nil {
		return nil, err
	}
	comment, ok := item.(*hnapi.Comment)
	if !ok {
		return nil, errors.New("not a comment")
	}
	return comment, nil
}

func (m StoryModel) incCommentSelection() int {
	return (m.currentSelection + 1) % len(m.comments)
}

func (m StoryModel) decCommentSelection() int {
	if m.currentSelection-1 < 0 {
		return len(m.comments) - 1
	}
	return m.currentSelection - 1
}
