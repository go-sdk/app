// Package app 提供 core、database 和 server 的约定式应用运行时。
package app

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/lifex"
	"github.com/go-sdk/core/logx"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/database/dbx"
	"github.com/go-sdk/database/dbx/migrate"
	"github.com/go-sdk/server/standard"
)

var defaultServer atomic.Pointer[standard.Server]

// Server 返回由 Run 初始化的进程级 Server 实例。
func Server() *standard.Server {
	server := defaultServer.Load()
	if server == nil {
		osx.Panic("app: server is not initialized")
	}
	return server
}

// Run 按固定顺序初始化数据库、迁移、业务数据和 Server，并等待应用退出。
// 每个进程只能调用一次；所有 Register 调用必须先于 Run 完成。
func Run() error {
	registered := registry.freeze()
	value, err := loadSettings()
	if err != nil {
		return err
	}

	logx.Init(config.MustGet("log.file.path", ""))
	defer logx.Close()
	logx.SetGlobalKV("service", value.name)
	logx.SetGlobalKV("version", osx.GetVersion().Version)

	db, err := openDatabase(value)
	if err != nil {
		return errx.Wrap(err, "initialize database")
	}
	if err = registerInitializers(db, registered); err != nil {
		return shutdown(err)
	}

	options, err := serverOptions(value, registered)
	if err != nil {
		return shutdown(err)
	}
	server, err := standard.New(options...)
	if err != nil {
		return shutdown(errx.Wrap(err, "initialize server"))
	}
	for _, item := range registered.routes {
		if err = server.HandlePath(item.method, item.path, item.handler); err != nil {
			return shutdown(errx.Wrapf(err, "register route %s %s", item.method, item.path))
		}
	}
	defaultServer.Store(server)

	if err = lifex.Init(); err != nil {
		return shutdown(err)
	}
	return lifex.Wait()
}

// Main 运行应用，并将非正常退出转换为进程退出码 1。
func Main() {
	if err := Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func registerInitializers(db *dbx.DB, registered registrationSnapshot) error {
	if len(registered.migrations) > 0 {
		migrator, err := migrate.New(db, registered.migrations, registered.migrationOptions...)
		if err != nil {
			return errx.Wrap(err, "initialize migrations")
		}
		lifex.OnInit(func() error { return migrator.Up(context.Background()) })
	}
	for _, bootstrap := range registered.bootstraps {
		if bootstrap == nil {
			return errx.New("bootstrap function must not be nil")
		}
		fn := bootstrap
		lifex.OnInit(func() error { return fn(context.Background(), db) })
	}
	return nil
}

func serverOptions(value settings, registered registrationSnapshot) ([]standard.Option, error) {
	options := []standard.Option{
		standard.WithName(value.name),
		standard.WithAddress(value.serverAddress),
	}
	if value.jwtSecret != "" {
		options = append(options, standard.WithJWTSecret([]byte(value.jwtSecret)))
	}
	if value.serverReflection {
		options = append(options, standard.WithReflection())
	}
	if len(registered.grpcRegisters) > 0 {
		options = append(options, standard.WithGRPCRegister(registered.grpcRegisters...))
	}
	if len(registered.gatewayRegisters) > 0 {
		options = append(options, standard.WithGatewayRegister(registered.gatewayRegisters...))
	}
	options = append(options, registered.serverOptions...)
	for _, factory := range registered.serverOptionFactories {
		if factory == nil {
			return nil, errx.New("server option factory must not be nil")
		}
		option, err := factory()
		if err != nil {
			return nil, errx.Wrap(err, "create server option")
		}
		if option == nil {
			return nil, errx.New("server option factory returned nil")
		}
		options = append(options, option)
	}
	return options, nil
}

func shutdown(reason error) error {
	lifex.Shutdown(reason)
	return lifex.Wait()
}
