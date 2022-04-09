package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/go-hn/pkg/hnapi"
	"github.com/hoenn/whohirin/pkg/data"
)

type App struct {
	current tea.Model
	fetcher *data.Fetcher
}

const defaultWidth = 80

func NewApp(userID string, client *hnapi.HNClient) (*App, error) {
	f, err := data.NewFetcher(userID)
	if err != nil {
		return nil, fmt.Errorf("could not initialize fetcher: %w", err)
	}
	return &App{
		current: *NewPostList(f),
		fetcher: f,
	}, nil
}

var initWindowSize *tea.WindowSizeMsg

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if initWindowSize == nil {
			initWindowSize = &msg
		}
	}
	return a.current.Update(msg)
}
func (a App) View() string {
	return a.current.View()
}
func (a App) Init() tea.Cmd {
	return nil
}
