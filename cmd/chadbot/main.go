package main

import (
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/tui"
	"github.com/ziemianek/chadbot/internal/twitch"
)

func main() {
	repo := twitch.NewFileSecretRepo("build/.token.json")
	client, err := twitch.NewClient(repo)
	if err != nil {
		log.Fatal(err)
	}
	app := tui.NewApp(client, true)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	log.Info("chadbot running!")
}
