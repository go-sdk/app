package app

import (
	"strings"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/osx"
)

type settings struct {
	name             string
	databaseDriver   string
	databaseDSN      string
	serverAddress    string
	jwtSecret        string
	serverReflection bool
}

func loadSettings() (settings, error) {
	value := settings{
		name:             strings.TrimSpace(config.MustGet("app.name", osx.ExeName())),
		databaseDriver:   strings.TrimSpace(config.MustGet("database.driver", "postgres")),
		databaseDSN:      strings.TrimSpace(config.MustGet("database.dsn", "")),
		serverAddress:    strings.TrimSpace(config.MustGet("server.address", ":8080")),
		jwtSecret:        config.MustGet("auth.jwt_secret", ""),
		serverReflection: config.MustGet("server.reflection", false),
	}
	if value.name == "" {
		return settings{}, errx.New("app name must not be empty")
	}
	if value.databaseDriver == "" {
		return settings{}, errx.New("database driver must not be empty")
	}
	if value.databaseDSN == "" {
		return settings{}, errx.New("database dsn must not be empty")
	}
	if value.serverAddress == "" {
		return settings{}, errx.New("server address must not be empty")
	}
	return value, nil
}
