package config

type Config struct {
	Address  string `yaml:"address,omitempty"`
	UseTLS   bool   `yaml:"useTLS,omitempty"`
	TLSKey   string `yaml:"tlsKey,omitempty"`
	TLSCert  string `yaml:"tlsCert,omitempty"`
	Tickrate int    `yaml:"tickrate,omitempty"`
}
