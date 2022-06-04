package server

import (
	"github.com/danny-m08/music-match/logging"
	"net/http"

	"github.com/danny-m08/music-match/config"
)

//StartServer runs a http server using the given config object
func StartServer(conf config.HTTPConfig) error {
	for pattern, handler := range handlers {
		http.HandleFunc(pattern, handler)
	}

	logging.Info("Server starting and listening on " + conf.ListenAddr)
	if conf.TLS != nil && conf.TLS.Enabled {
		logging.Debug("TLS enabled")
		err := http.ListenAndServeTLS(conf.ListenAddr, conf.TLS.CertFile, conf.TLS.KeyFile, nil)
		if err != nil {
			return err
		}
	}
	logging.Debug("TLS disabled")

	return http.ListenAndServe(conf.ListenAddr, nil)
}
