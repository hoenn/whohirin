package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoenn/go-hn/pkg/hnapi"
)

type App struct {
	current tea.Model
}

func NewApp(userID string, client *hnapi.HNClient) (*App, error) {
	ids, err := getPostList(client, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get post list for userID %s: %w", userID, err)
	}
	return &App{
		current: *NewStoryList(ids, client),
	}, nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a.current.Update(msg)
}
func (a App) View() string {
	return a.current.View()
}
func (a App) Init() tea.Cmd {
	return nil
}
func getPostList(client *hnapi.HNClient, userID string) ([]int, error) {
	u, err := client.User(userID)
	if err != nil {
		return []int{}, err
	}
	return u.Submitted, nil
}
