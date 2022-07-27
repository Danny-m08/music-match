package main

import (
	"fmt"
	"github.com/danny-m08/music-match/logging"
	"os"
	"os/signal"
	"syscall"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/server"
)

const configFile = "config.yaml"

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err := config.CreateConfigFromFile(configFile)
	if err != nil {
		logging.Error("Unable to create config from file: " + err.Error())
		os.Exit(1)
	}

	serv, err := server.NewServer(config.GetGlobalConfig().GetHTTPServerConfig(), config.GetGlobalConfig().GetDBConfig())
	if err != nil {
		logging.Error("Unable to start server: " + err.Error())
		os.Exit(1)
	}

	err = serv.StartServer()
	if err != nil {
		logging.Error("Unable to start HTTP Server: " + err.Error())
		os.Exit(1)
	}

	sig := <-sigChan
	logging.Info(fmt.Sprintf("Signal %s caught -- terminating program", sig.String()))
	if serv.Close() != nil {
		logging.Error("Error closing server: " + err.Error())
	}
	os.Exit(0)

}
