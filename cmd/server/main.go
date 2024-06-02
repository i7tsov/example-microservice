package main

import (
	"fmt"
	"strings"

	"github.com/i7tsov/example-microservice/internal/server"
	"github.com/i7tsov/example-microservice/pkg/graceful"
	"github.com/i7tsov/example-microservice/pkg/repository/users"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Overriden with repository tag+hash version by build script.
var version = "dev"

const appName = "service"

func main() {
	// Initialization code is moved to "run" function in order to
	// enable any deferred cleanup to finish before service exits
	// (Fatal calls os.Exit)
	err := run()
	if err != nil {
		logrus.Fatalf("Fatal error: %v", err)
	}

	logrus.Info("Server stopped gracefully")
}

func run() error {
	// Viper status is saved in order to log it after the logger initializes.
	// Logger needs log level from viper, so viper initializes first.
	viperStatus := initViper()

	// We're using logrus as a logger. While it's a singleton, it's much
	// easier to use in code when accepted as a primary logger for the service without
	// passing it around many components, most of which need logger.
	//
	// Changing logger will require significant modification of the code,
	// but simplicity of use overcomes possible risks of changing every logging
	// line in the code (which are mostly routine changes).
	initLogger()

	// It's nice to have server version at the start of the logs.
	// Improvement may include printing version when binary is run
	// with a special command, e.g. "server version", if applicable.
	logrus.Infof("Server %v version %v", appName, version)
	logrus.Info(viperStatus)
	logrus.Infof("Logging level: %v", logrus.StandardLogger().Level)

	// Creating database access component.
	var dbCfg users.Config
	err := viper.Sub(dbConfig).Unmarshal(&dbCfg)
	if err != nil {
		return fmt.Errorf("reading database config: %w", err)
	}
	users, err := users.New(dbCfg)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// Creating HTTP server component.
	var srvCfg server.Config
	err = viper.Sub(serverConfig).Unmarshal(&srvCfg)
	if err != nil {
		return fmt.Errorf("reading database config: %w", err)
	}
	srv, err := server.New(
		srvCfg,
		server.Dependencies{
			UsersRepo: users,
		},
	)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	// This runs server listening routine and catches OS signals
	// to perform graceful termination by cancelling context
	// passed to srv.Serve.
	//
	// In real world scenario there are several long-lived goroutines,
	// and when any of them returns an error, that means the service can't function
	// anymore. Graceful package assures all the routines are gracefully stopped
	// and the error that caused termination is returned.
	return graceful.Run(graceful.Config{}, srv.Serve)
}

func initViper() (status string) {
	viper.AddConfigPath(".")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v/", appName))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		status = "Config file not found, using environment variables and/or flags"
	} else {
		status = "Loaded config file"
	}

	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	return
}

func initLogger() {
	level := viper.GetString(logLevel)
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		ll = logrus.InfoLevel
		logrus.StandardLogger().Warnf("Invalid log level %v, set to info", level)
	} else {
		logrus.SetLevel(ll)
	}
}
