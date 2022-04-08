package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/hoenn/go-hn/pkg/hnapi"

	tea "github.com/charmbracelet/bubbletea"
)

type Client struct {
	hn *hnapi.HNClient
}

const user = "whoishiring"

func main() {
	client := &Client{
		hn: hnapi.NewHNClient(),
	}
	// get a list of story IDs
	ids, err := client.getPostList(user)
	if err != nil {
		fmt.Printf("could not get post list: %s", err)
		return
	}

	m := model{
		storyTitles:      make(map[int]string),
		stories:          ids,
		currentSelection: 0,
		client:           client,
	}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("oops: %v", err)
		os.Exit(1)
	}

}

func (c *Client) getPostList(userID string) ([]int, error) {
	u, err := c.hn.User(userID)
	if err != nil {
		return []int{}, err
	}
	return u.Submitted, nil
}

func (c *Client) getStoryTitle(id int) (string, error) {
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

type model struct {
	storyTitles      map[int]string // lazy load titles
	stories          []int
	currentSelection int
	client           *Client
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	storyID := m.stories[m.currentSelection]
	title, ok := m.storyTitles[storyID]
	if title == "" || !ok {
		t, err := m.client.getStoryTitle(storyID)
		if err != nil {
			return fmt.Sprintf("error getting title... %v", err)
		}
		m.storyTitles[storyID] = t
		title = t
	}
	return title
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.currentSelection = decWithWrap(m.currentSelection, 1, len(m.stories))
		case "right", "l":
			m.currentSelection = incWithWrap(m.currentSelection, 1, len(m.stories))
		}
	}
	return m, nil
}

func incWithWrap(i, inc, max int) int {
	return (i + inc) % max
}

func decWithWrap(i, dec, max int) int {
	if i-dec < 0 {
		return max - 1
	}
	return i - dec
}
