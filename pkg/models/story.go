package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/go-hn/pkg/hnapi"
)

type StoryModel struct {
	storyID  int
	hn       *hnapi.HNClient
	previous tea.Model
}

func NewStory(storyID int, previous tea.Model, client *hnapi.HNClient) *StoryModel {
	return &StoryModel{
		storyID:  storyID,
		hn:       client,
		previous: previous,
	}
}

func (m StoryModel) Init() tea.Cmd {
	return nil
}

func (m StoryModel) View() string {
	return fmt.Sprintf("you are now viewing %d", m.storyID)
}

func (m StoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			return m.previous, nil
		}
	}
	return m, nil
}
