// Package config loads server settings from command-line flags and environment variables.
package config

import (
	"flag"
	"os"
)

const (
	DefaultAddr   = "127.0.0.1:5000"
	DefaultDBPath = "test.db"
)

type Config struct {
	Addr   string
	DBPath string
}

// Load parses args (without the program name). Flags take precedence over
// the ADDR and DB_PATH environment variables.
func Load(args []string) (Config, error) {
	var cfg Config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", envOr("ADDR", DefaultAddr), "address to listen on")
	fs.StringVar(&cfg.DBPath, "db", envOr("DB_PATH", DefaultDBPath), "path to the SQLite database file")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
