package app

import (
	"strings"
	"time"

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
	Enabled          bool          `json:"enabled"`
	Mode             rdx.Mode      `json:"mode"`
	Addresses        []string      `json:"addresses"`
	MasterName       string        `json:"master_name"`
	Username         string        `json:"username"`
	Password         string        `json:"password"`
	SentinelUsername string        `json:"sentinel_username"`
	SentinelPassword string        `json:"sentinel_password"`
	Database         int           `json:"database"`
	ClientName       string        `json:"client_name"`
	DialTimeout      time.Duration `json:"dial_timeout"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	PoolTimeout      time.Duration `json:"pool_timeout"`
	PoolSize         int           `json:"pool_size"`
	MinIdleConns     int           `json:"min_idle_conns"`
	MaxIdleConns     int           `json:"max_idle_conns"`
	MaxActiveConns   int           `json:"max_active_conns"`
	ConnMaxIdleTime  time.Duration `json:"conn_max_idle_time"`
	ConnMaxLifetime  time.Duration `json:"conn_max_lifetime"`
	TLS              rdx.TLSConfig `json:"tls"`
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
			Mode:      rdx.ModeStandalone,
			Addresses: []string{"127.0.0.1:6379"},
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
	value.Redis.MasterName = strings.TrimSpace(value.Redis.MasterName)
	value.Redis.ClientName = strings.TrimSpace(value.Redis.ClientName)
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

func (s redisSettings) config() rdx.Config {
	return rdx.Config{
		Mode:             s.Mode,
		Addresses:        s.Addresses,
		MasterName:       s.MasterName,
		Username:         s.Username,
		Password:         s.Password,
		SentinelUsername: s.SentinelUsername,
		SentinelPassword: s.SentinelPassword,
		Database:         s.Database,
		ClientName:       s.ClientName,
		DialTimeout:      s.DialTimeout,
		ReadTimeout:      s.ReadTimeout,
		WriteTimeout:     s.WriteTimeout,
		PoolTimeout:      s.PoolTimeout,
		PoolSize:         s.PoolSize,
		MinIdleConns:     s.MinIdleConns,
		MaxIdleConns:     s.MaxIdleConns,
		MaxActiveConns:   s.MaxActiveConns,
		ConnMaxIdleTime:  s.ConnMaxIdleTime,
		ConnMaxLifetime:  s.ConnMaxLifetime,
		TLS:              s.TLS,
	}
}
