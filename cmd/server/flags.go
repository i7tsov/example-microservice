package main

import (
	"github.com/spf13/pflag"
)

const (
	logLevel   = "log.level"
	address    = "server.address"
	port       = "server.port"
	dbHost     = "db.host"
	dbPort     = "db.port"
	dbUser     = "db.user"
	dbPassword = "db.password"
	dbName     = "db.name"

	dbConfig     = "db"
	serverConfig = "server"
)

// Set up command line flags.
func init() {
	_ = pflag.String(logLevel, "info", "logging level (trace, debug, info, warning, error, critical)")
	_ = pflag.String(address, "", "address for server to listen to")
	_ = pflag.String(port, "10002", "port for server to listen to")
	_ = pflag.String(dbHost, "localhost", "host of the database")
	_ = pflag.String(dbPort, "5432", "port of the database")
	_ = pflag.String(dbUser, "userman", "database user")
	_ = pflag.String(dbPassword, "", "database user password")
	_ = pflag.String(dbName, "users", "database name")

	pflag.Parse()
}
