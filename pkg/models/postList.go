package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/whohirin/pkg/data"
)

type postListModel struct {
	keys             postListKeyMap
	help             help.Model
	postIDs          []string
	currentSelection int
	hn               *data.Fetcher
}

func NewPostList(f *data.Fetcher) *postListModel {
	postIDs := f.PostList()
	h := help.New()
	h.Width = defaultWidth
	return &postListModel{
		keys:             postListKeys,
		help:             h,
		postIDs:          postIDs,
		currentSelection: 0,
		hn:               f,
	}
}

type postListKeyMap struct {
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Quit   key.Binding
	Help   key.Binding
}

func (k postListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}
func (k postListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Select},
		{k.Help, k.Quit},
	}
}

var postListKeys = postListKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("← / h", "previous"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→ / l", "next"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter / space", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q / ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h / ?", "toggle help"),
	),
}

func (m postListModel) Init() tea.Cmd {
	return nil
}

func (m postListModel) View() string {
	postID := m.postIDs[m.currentSelection]
	p, err := m.hn.Post(postID)

	output := p.Title
	if err != nil || p == nil {
		output = fmt.Sprint("could not find post, it may have been [deleted]")
	}
	helpView := m.help.View(m.keys)
	height := 8 - strings.Count(output, "\n") - strings.Count(helpView, "\n")

	return "\n" + output + strings.Repeat("\n", height) + helpView
}

func (m postListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.currentSelection = m.decCommentSelection()
		case key.Matches(msg, m.keys.Right):
			m.currentSelection = m.incCommentSelection()
		case key.Matches(msg, m.keys.Select):
			s, err := NewPost(m.postIDs[m.currentSelection], m, m.hn)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to show post: %w", err))
				return m, nil
			}
			return s, nil
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	}
	return m, nil
}

func (m postListModel) incCommentSelection() int {
	return (m.currentSelection + 1) % len(m.postIDs)
}

func (m postListModel) decCommentSelection() int {
	if m.currentSelection-1 < 0 {
		return len(m.postIDs) - 1
	}
	return m.currentSelection - 1
}
