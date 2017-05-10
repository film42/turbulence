package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var config *Config

func initConfigFromFlags() {
	configPtr := flag.String("config", "", "config file")
	portPtr := flag.Int("port", 25000, "listen port")
	stripProxyHeadersPtr := flag.Bool("strip-proxy-headers", true, "strip proxy headers from http requests")
	usernamePtr := flag.String("username", "", "username for proxy authentication")
	passwordPtr := flag.String("password", "", "password for proxy authentication")
	shutdownTimeoutPtr := flag.Int("shutdown-timeout", 60, "seconds to wait while cleaning up for connections to finish")
	flag.Parse()

	config = &Config{
		Port:              *portPtr,
		StripProxyHeaders: *stripProxyHeadersPtr,
		ShutdownTimeout:   *shutdownTimeoutPtr,
	}

	if *usernamePtr != "" {
		config.Credentials = []Credential{
			{Username: *usernamePtr, Password: *passwordPtr},
		}
	}

	if *configPtr != "" {
		configFile, err := os.Open(*configPtr)
		if err != nil {
			logger.Fatal.Println("Could not open config file", err)
			os.Exit(1)
		}

		err = config.LoadConfigFromReader(configFile)
		if err != nil {
			logger.Fatal.Println("Could not parse config file", err)
			os.Exit(1)
		}

		configFile.Close()
	}
}

func main() {
	InitLogger()
	initConfigFromFlags()

	err := config.Validate()
	if err != nil {
		logger.Fatal.Println("Config is not valid:", err)
		os.Exit(1)
	}

	if config.AuthenticationRequired() {
		logger.Info.Println("Credentials provided. Proxy authentication will be required for all connections.")
	}

	logger.Info.Println("Prepare for takeoff...")

	server := NewServer()
	go server.ListenAndServe(config.Port)

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	logger.Info.Println(<-signalChannel, "detected! Waiting for connections to finish. Interrupt again to force stop.")

	shutdownCompleteChannel := make(chan bool)
	go func() {
		server.Shutdown()
		shutdownCompleteChannel <- true
	}()

	logger.Info.Println("Waiting up to", config.ShutdownTimeoutDuration(), "for open connections to finish.")
	shutdownTimer := time.NewTimer(config.ShutdownTimeoutDuration())

	select {
	case <-shutdownTimer.C:
		logger.Warn.Println("Graceful shutdown timeout expired. Forcing shutdown!")
	case <-shutdownCompleteChannel:
		logger.Info.Println("Graceful shutdown copmlete!")
	case <-signalChannel:
		logger.Info.Println("Force shutdown complete!")
	}

}
