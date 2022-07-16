package config

type Config struct {
	DB   *Neo4jConfig `yaml:"neo4j"`
	Http *HTTPConfig  `yaml:"http"`
}

//HTTPConfig provisions an http server from the given config
type HTTPConfig struct {
	ListenAddr string     `yaml:"listen-address"`
	TLS        *TLSConfig `yaml:"tls,omitempty"`
}

//TLSConfig provides TLS configuration options for the server
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert-path"`
	KeyFile  string `yaml:"key-path"`
	CAFile   string `yaml:"ca-path"`
}

type Neo4jConfig struct {
	URI         string       `yaml:"endpoint"`
	Database    string       `yaml:"database,omitempty"`
	Plaintext   bool         `yaml:"plaintext,omitempty"`
	Username    string       `yaml:"username,omitempty"`
	Password    string       `yaml:"password,omitempty"`
	BatchSize   int          `yaml:"batch-size,omitempty"`
	RetryConfig *RetryConfig `yaml:"retry,omitempty"`
}

type RetryConfig struct {
	//max      uint `yaml:"max"`
	//interval uint `yaml:"max"`
	//timeout  uint `yaml:"timeout"`
}
