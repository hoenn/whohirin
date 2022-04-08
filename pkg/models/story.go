package models

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/go-hn/pkg/hnapi"
)

type StoryModel struct {
	storyID          int
	hn               *hnapi.HNClient
	previous         tea.Model
	comments         []int
	content          string
	commentContent   map[int]string // lazy loaded
	currentSelection int
}

func NewStory(storyID int, previous tea.Model, client *hnapi.HNClient) *StoryModel {
	item, err := client.Item(fmt.Sprintf("%d", storyID))
	if err != nil {
		panic(err)
	}
	story, ok := item.(*hnapi.Story)
	if !ok {
		panic(errors.New("fix this"))
	}
	return &StoryModel{
		storyID:          storyID,
		hn:               client,
		previous:         previous,
		comments:         story.Kids,
		commentContent:   make(map[int]string),
		currentSelection: 0,
	}
}

func (m StoryModel) Init() tea.Cmd {
	return nil
}

func (m StoryModel) View() string {
	commentID := m.comments[m.currentSelection]
	content, ok := m.commentContent[commentID]
	if content == "" || !ok {
		cc, err := m.getCommentContent(commentID)
		if err != nil {
			return fmt.Sprintf("error getting comment... %v", err)
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
			m.currentSelection = decSelectionWithWrap(m.currentSelection, 1, len(m.comments))
		case "right", "l":
			m.currentSelection = incSelectionWithWrap(m.currentSelection, 1, len(m.comments))
		case "backspace":
			return m.previous, nil

		}

	}
	return m, nil
}

func (c *StoryModel) getCommentContent(id int) (string, error) {
	item, err := c.hn.Item(fmt.Sprintf("%d", id))
	if err != nil {
		return "", err
	}
	comment, ok := item.(*hnapi.Comment)
	if !ok {
		return "", errors.New("not a comment")
	}
	return comment.Text, nil
}
