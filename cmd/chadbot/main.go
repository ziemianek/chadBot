package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/tui"
)

func main() {
	var app *tui.App = tui.NewApp(true)
	if err := app.Run(); err != nil {
		log.Errorf("Could not start the app: %v", err)
		os.Exit(1)
	}
}
