package main

import (
	app "github.com/rutkin/gofermart/internal"
	"github.com/rutkin/gofermart/internal/config"
	"github.com/rutkin/gofermart/internal/logger"
)

func main() {
	config := config.MakeConfig()

	err := logger.Initialize(config.LogLevel)
	if err != nil {
		panic(err)
	}

	server, err := app.MakeServer(config)
	if err != nil {
		panic(err)
	}
	server.Start()
}
