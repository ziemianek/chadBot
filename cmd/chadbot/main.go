package main

import (
	"github.com/joho/godotenv"
	"github.com/ziemianek/chadbot/internal/logger"
)

func main() {
	logger := logger.New(true) // true means debug mode on, set false for debug off

	// cleaner error checking
	check := func(err error) {
		if err != nil {
			logger.Error(err)
		}
	}

	err := godotenv.Load()
	check(err)

	logger.Info("Welcome to ChadBot")
}
