package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadingConfigFromFile(t *testing.T) {
	sampleConfigFile, _ := os.Open("test/config.json")
	defer sampleConfigFile.Close()

	sampleConfig, _ := NewConfigFromReader(sampleConfigFile)
	if err := sampleConfig.Validate(); err != nil {
		t.Fatal("Expected config to be valid but found: ", err)
	}

	if sampleConfig.Port != 26000 {
		t.Fatal("Expected port to be 26000 but found:", sampleConfig.Port)
	}

	if !sampleConfig.StripProxyHeaders {
		t.Fatal("Expected sampleConfig.StripProxyHeaders to be true")
	}

	if !sampleConfig.AuthenticationRequired() {
		t.Fatal("Expected sample config to require authentication")
	}

	if sampleConfig.ShutdownTimeout != 100 {
		t.Fatal("Expected sample config shutdown timeout to be 100 but found", sampleConfig.ShutdownTimeout)
	}

	expectedShutdownDuration := time.Duration(100 * time.Second)
	if sampleConfig.ShutdownTimeoutDuration() != expectedShutdownDuration {
		t.Fatal("Expected sample config shutdown timeout to be", expectedShutdownDuration, "but found", sampleConfig.ShutdownTimeoutDuration())
	}

	credentialsLen := len(sampleConfig.Credentials)
	if credentialsLen != 2 {
		t.Fatal("Expected config to have 2 credentials but found:", credentialsLen)
	}

	credential := Credential{Username: "login", Password: "yolo"}
	if sampleConfig.Credentials[0] != credential {
		t.Fatal("Expected", credential, "but found", sampleConfig.Credentials[0])
	}

	credential = Credential{Username: "ron.swanson", Password: "g0ld5topsTheGovt"}
	if sampleConfig.Credentials[1] != credential {
		t.Fatal("Expected", credential, "but found", sampleConfig.Credentials[1])
	}
}

func TestInvalidConfigFormat(t *testing.T) {
	invalidFormatReader := strings.NewReader(`{"testing":"ok...`)
	_, err := NewConfigFromReader(invalidFormatReader)
	if err == nil {
		t.Fatal("Config should have returned an error for being invalid json")
	}
}

func TestConfigInavlidUsername(t *testing.T) {
	sampleConfig := &Config{
		Credentials: []Credential{{Username: "", Password: "test"}},
	}

	if sampleConfig.Validate() != InvalidCredentials {
		t.Fatal("Expected username to be invalid")
	}
}

func TestConfigInavlidPassword(t *testing.T) {
	sampleConfig := &Config{
		Credentials: []Credential{{Username: "test", Password: ""}},
	}

	if sampleConfig.Validate() != InvalidCredentials {
		t.Fatal("Expected password to be invalid")
	}
}

func TestLoadConfigFromReader(t *testing.T) {
	sampleConfig := &Config{
		StripProxyHeaders: true,
	}

	if !sampleConfig.StripProxyHeaders {
		t.Fatal("Expected stip proxy headers to be true")
	}

	json := `{"strip_proxy_headers":false}`
	sampleConfig.LoadConfigFromReader(strings.NewReader(json))

	if sampleConfig.StripProxyHeaders {
		t.Fatal("Expected stip proxy headers to be false")
	}
}
