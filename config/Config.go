package config

// Config provides a structure that contains configurable options for the game server.
type Config struct {
	Address  string `yaml:"address,omitempty"`
	UseTLS   bool   `yaml:"useTLS,omitempty"`
	TLSKey   string `yaml:"tlsKey,omitempty"`
	TLSCert  string `yaml:"tlsCert,omitempty"`
	Tickrate int    `yaml:"tickrate,omitempty"`
}
