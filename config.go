package app

import (
	"strings"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/database/rdx"
)

type settings struct {
	App      appSettings      `json:"app"`
	Database databaseSettings `json:"database"`
	Redis    redisSettings    `json:"redis"`
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

type redisSettings struct {
	Enabled    bool `json:"enabled"`
	rdx.Config `json:",squash"`
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
		Redis: redisSettings{
			Config: rdx.Config{
				Mode:      rdx.ModeStandalone,
				Addresses: []string{"127.0.0.1:6379"},
			},
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
	value.Redis.Config.MasterName = strings.TrimSpace(value.Redis.Config.MasterName)
	value.Redis.Config.ClientName = strings.TrimSpace(value.Redis.Config.ClientName)
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
