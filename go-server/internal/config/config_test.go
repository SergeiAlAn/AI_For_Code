package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("DB_PATH", "")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != DefaultAddr || cfg.DBPath != DefaultDBPath {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadEnvAndFlagPrecedence(t *testing.T) {
	t.Setenv("ADDR", "0.0.0.0:8080")
	t.Setenv("DB_PATH", "env.db")

	cfg, err := Load([]string{"-db", "flag.db"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "0.0.0.0:8080" || cfg.DBPath != "flag.db" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadUnknownFlag(t *testing.T) {
	if _, err := Load([]string{"-nope"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
