package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/server"
)

const configFile = "config.yaml"

func main() {
	errChan := make(chan error)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	err := config.CreateConfigFromFile(configFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	go func() {
		errChan <- server.StartServer(config.GetGlobalConfig().GetHTTPServerConfig())
	}()

	select {
	case err := <-errChan:
		fmt.Println(err.Error())
		os.Exit(1)
	case sig := <-sigChan:
		fmt.Printf("Signal %s caught -- terminating program", sig.String())
		os.Exit(0)
	}
}
