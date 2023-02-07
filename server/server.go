package server

import (
	"errors"
	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/neo4j"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"regexp"
)

type server struct {
	neo4jClient *neo4j.Client
	httpConfig  *config.HTTPConfig
}

func NewServer(conf *config.HTTPConfig, neo4jConfig *config.Neo4jConfig) (*server, error) {
	client, err := neo4j.NewClient(neo4jConfig)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, errors.New("Http config cannot be nil")
	}

	return &server{
		neo4jClient: client,
		httpConfig:  conf,
	}, nil
}

// StartServer runs a http server using the given config object
func (s *server) StartServer() error {

	apiServer := http.NewServeMux()

	apiServer.HandleFunc("/login", s.login)
	apiServer.HandleFunc("/signup", s.newUser)
	apiServer.HandleFunc("/follow", s.follow)
	apiServer.HandleFunc("/followers", s.getFollowers)

	logging.Info("Server starting and listening on " + s.httpConfig.ListenAddr)
	if s.httpConfig.TLS != nil && s.httpConfig.TLS.Enabled {
		logging.Debug("TLS enabled")
		err := http.ListenAndServeTLS(s.httpConfig.ListenAddr, s.httpConfig.TLS.CertFile, s.httpConfig.TLS.KeyFile, apiServer)
		if err != nil {
			return err
		}
	}

	return http.ListenAndServe(s.httpConfig.ListenAddr, nil)
}

func (s *server) StartMetrics() error {
	metricsServer := http.NewServeMux()
	metricsServer.Handle("/metrics", promhttp.Handler())
	logging.Info("Metrics listening on " + s.httpConfig.MetricsAddr)

	return http.ListenAndServe(s.httpConfig.MetricsAddr, metricsServer)
}

func (s *server) Close() error {
	if s.neo4jClient != nil {
		return s.neo4jClient.Close()
	}

	return nil
}

func verifyEmail(email string) (bool, error) {
	return regexp.Match("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", []byte(email))
}
