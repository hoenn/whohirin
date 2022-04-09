package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/whohirin/pkg/data"
)

type postListModel struct {
	postIDs          []string
	currentSelection int
	hn               *data.Fetcher
}

func NewPostList(f *data.Fetcher) *postListModel {
	postIDs := f.PostList()
	return &postListModel{
		postIDs:          postIDs,
		currentSelection: 0,
		hn:               f,
	}
}

func (m postListModel) Init() tea.Cmd {
	return nil
}

func (m postListModel) View() string {
	// post Selection View
	postID := m.postIDs[m.currentSelection]
	p, err := m.hn.Post(postID)
	if err != nil || p == nil {
		return fmt.Sprint("could not find post, it may have been [deleted]")
	}
	return p.Title
}

func (m postListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.currentSelection = m.decCommentSelection()
		case "right", "l":
			m.currentSelection = m.incCommentSelection()
		case "enter":
			s, err := NewPost(m.postIDs[m.currentSelection], m, m.hn)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to show post: %w", err))
				return m, nil
			}
			return s, nil
		}
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
