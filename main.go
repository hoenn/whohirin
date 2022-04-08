package main

import (
	"fmt"
	"os"

	"github.com/hoenn/go-hn/pkg/hnapi"
	"github.com/hoenn/whohirin/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
)

const user = "whoishiring"

func main() {
	client := hnapi.NewHNClient()

	app, err := models.NewApp(user, client)
	if err != nil {
		fmt.Printf("Could not start application: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(app, tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("oops: %v", err)
		os.Exit(1)
	}

}
