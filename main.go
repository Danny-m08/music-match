package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/server"
)

var configFile string

const defaultConfig = "config.yaml"

func init() {
	configFile = os.Getenv("CONFIG_PATH")
	if configFile == "" {
		configFile = defaultConfig
	}

	println(configFile)
}

func main() {
	errChan := make(chan error)
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

	go func() {
		errChan <- serv.StartServer()
	}()

	go func() {
		errChan <- serv.StartMetrics()
	}()

	select {
	case err := <-errChan:
		logging.Error(err.Error())
		os.Exit(1)
	case sig := <-sigChan:
		logging.Info(fmt.Sprintf("Signal %s caught -- terminating program", sig.String()))
		if serv.Close() != nil {
			logging.Error("Error closing server: " + err.Error())
		}
		os.Exit(0)
	}
}
