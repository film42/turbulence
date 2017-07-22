package main

import (
	"encoding/json"
	"errors"
	"io"
)

var InvalidCredentials = errors.New("Invalid credentials provided. Must have a username/ password or none at all.")

type Credential struct {
	Username string
	Password string
}

type Config struct {
	Credentials          []Credential
	StripProxyHeaders    bool `json:"strip_proxy_headers"`
	Port                 int
	UseIncomingLocalAddr bool `json:"use_incoming_local_addr"`
}

func (config *Config) AuthenticationRequired() bool {
	return len(config.Credentials) > 0
}

func validCredentials(username, password string) bool {
	if username == "" && password == "" {
		return true
	}
	if username != "" && password != "" {
		return true
	}
	return false
}

func (config *Config) Validate() error {
	for _, credential := range config.Credentials {
		if !validCredentials(credential.Username, credential.Password) {
			return InvalidCredentials
		}
	}

	return nil
}

func (config *Config) IsAuthenticated(username, password string) bool {
	for _, credential := range config.Credentials {
		if credential.Username == username && credential.Password == password {
			return true
		}
	}

	return false
}

func NewConfigFromReader(reader io.Reader) (*Config, error) {
	config := new(Config)
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
