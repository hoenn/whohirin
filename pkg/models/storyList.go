package models

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/go-hn/pkg/hnapi"
)

type StoryListModel struct {
	storyTitles      map[int]string // lazy load titles
	stories          []int
	storySelected    bool
	currentSelection int
	hn               *hnapi.HNClient
}

func NewStoryList(storyIDs []int, client *hnapi.HNClient) *StoryListModel {
	return &StoryListModel{
		storyTitles:      make(map[int]string),
		stories:          storyIDs,
		currentSelection: 0,
		storySelected:    false,
		hn:               client,
	}
}

func (m StoryListModel) Init() tea.Cmd {
	return nil
}

func (m StoryListModel) View() string {
	// Story Selection View
	storyID := m.stories[m.currentSelection]
	title, ok := m.storyTitles[storyID]
	if title == "" || !ok {
		t, err := m.getStoryTitle(storyID)
		if err != nil {
			return fmt.Sprintf("error getting title... %v", err)
		}
		m.storyTitles[storyID] = t
		title = t
	}
	return title
}

func (m StoryListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.decSelectionWithWrap(1)
		case "right", "l":
			m.incSelectionWithWrap(1)
		case "enter":
			return NewStory(m.stories[m.currentSelection], m, m.hn), nil
		}
	}
	return m, nil
}

func (c *StoryListModel) getStoryTitle(id int) (string, error) {
	item, err := c.hn.Item(fmt.Sprintf("%d", id))
	if err != nil {
		return "", err
	}
	story, ok := item.(*hnapi.Story)
	if !ok {
		return "", errors.New("not a story")
	}
	return story.Title, nil
}

func (m *StoryListModel) incSelectionWithWrap(inc int) {
	m.currentSelection = (m.currentSelection + inc) % len(m.stories)
}

func (m *StoryListModel) decSelectionWithWrap(dec int) {
	if m.currentSelection-dec < 0 {
		m.currentSelection = len(m.stories) - 1
	}
	m.currentSelection = m.currentSelection - 1
}
