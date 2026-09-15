package app

import (
	"strings"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/osx"
)

type settings struct {
	App      appSettings      `json:"app"`
	Database databaseSettings `json:"database"`
	Server   serverSettings   `json:"server"`
	Auth     authSettings     `json:"auth"`
}

type appSettings struct {
	Name string `json:"name"`
}

type databaseSettings struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

type serverSettings struct {
	Address    string `json:"address"`
	Reflection bool   `json:"reflection"`
}

type authSettings struct {
	JWTSecret string `json:"jwt_secret"`
}

func loadSettings() (settings, error) {
	value := settings{
		App: appSettings{
			Name: osx.ExeName(),
		},
		Database: databaseSettings{
			Driver: "postgres",
		},
		Server: serverSettings{
			Address: ":8080",
		},
	}
	if err := config.DecodeTo(&value); err != nil {
		return settings{}, errx.Wrap(err, "decode app config")
	}
	value.App.Name = strings.TrimSpace(value.App.Name)
	value.Database.Driver = strings.TrimSpace(value.Database.Driver)
	value.Database.DSN = strings.TrimSpace(value.Database.DSN)
	value.Server.Address = strings.TrimSpace(value.Server.Address)
	if value.App.Name == "" {
		return settings{}, errx.New("app name must not be empty")
	}
	if value.Database.Driver == "" {
		return settings{}, errx.New("database driver must not be empty")
	}
	if value.Database.DSN == "" {
		return settings{}, errx.New("database dsn must not be empty")
	}
	if value.Server.Address == "" {
		return settings{}, errx.New("server address must not be empty")
	}
	return value, nil
}
