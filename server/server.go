package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/neo4j"
)

type server struct {
	neo4jClient neo4j.Neo4jClient
	httpConfig  *config.HTTPConfig
	sessions    map[string]string
	dataStore   string
}

type Server struct {
	s server
}

func NewServer(conf *config.HTTPConfig, neo4jConfig *config.Neo4jConfig) (*server, error) {
	var client *neo4j.Client
	var err error

	client, err = neo4j.NewClient(neo4jConfig)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(conf.DataStore, 0777)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("Unable to create data store directory at %s: %s", conf.DataStore, err.Error())
	}

	if conf == nil {
		return nil, errors.New("Http config cannot be nil")
	}

	return &server{
		neo4jClient: client,
		httpConfig:  conf,
		sessions:    make(map[string]string),
		dataStore:   conf.DataStore,
	}, nil
}

//StartServer runs a http server using the given config object
func (s *server) StartServer() error {

	http.HandleFunc("/login", s.login)
	http.HandleFunc("/getUser", s.getUserInfo)
	http.HandleFunc("/signup", s.newUser)
	http.HandleFunc("/follow", s.follow)
	http.HandleFunc("/followers", s.getFollowers)
	http.HandleFunc("/upload", s.uploadFile)

	logging.Info("Server starting and listening on " + s.httpConfig.ListenAddr)
	if s.httpConfig.TLS != nil && s.httpConfig.TLS.Enabled {
		logging.Debug("TLS enabled")
		err := http.ListenAndServeTLS(s.httpConfig.ListenAddr, s.httpConfig.TLS.CertFile, s.httpConfig.TLS.KeyFile, nil)
		if err != nil {
			return err
		}
	}
	logging.Debug("TLS disabled")

	errChan := make(chan error)

	go func() {
		errChan <- http.ListenAndServe(s.httpConfig.ListenAddr, nil)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(2 * time.Second):
		return nil
	}
}

func (s *server) Close() error {
	if s.neo4jClient != nil {
		logging.Info("Closing neo4j Client...")
		return s.neo4jClient.Close()
	}

	return nil
}

func verifyEmail(email string) (bool, error) {
	return regexp.Match("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", []byte(email))
}
