package app

import (
	"context"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/database/dbx"
	"github.com/go-sdk/database/dbx/migrate"
	"github.com/go-sdk/server/standard"
)

// BootstrapFunc 在数据库迁移完成后、Server 启动前执行一次业务初始化。
type BootstrapFunc func(context.Context, *dbx.DB) error

// ServerOptionFactory 在数据库初始化后创建 Server Option。
// 需要访问 DB 的鉴权拦截器等组件应通过该工厂注册。
type ServerOptionFactory func() (standard.Option, error)

type route struct {
	method  string
	path    string
	handler standard.HandlerFunc
}

type registrations struct {
	mu                    sync.Mutex
	frozen                bool
	migrations            migrate.Migrations
	migrationOptions      []migrate.Option
	bootstraps            []BootstrapFunc
	grpcRegisters         []standard.GRPCRegisterFunc
	gatewayRegisters      []standard.GatewayRegisterFunc
	routes                []route
	serverOptions         []standard.Option
	serverOptionFactories []ServerOptionFactory
}

type registrationSnapshot struct {
	migrations            migrate.Migrations
	migrationOptions      []migrate.Option
	bootstraps            []BootstrapFunc
	grpcRegisters         []standard.GRPCRegisterFunc
	gatewayRegisters      []standard.GatewayRegisterFunc
	routes                []route
	serverOptions         []standard.Option
	serverOptionFactories []ServerOptionFactory
}

var registry registrations

// RegisterMigration 从直接调用方的文件名生成迁移 ID 并登记数据库迁移。
// 文件名必须符合 YYYYMMDD_HHMMSS_NN_description.go，发布后不得重命名。
func RegisterMigration(up, down func(*dbx.DB) error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		osx.Panic("app: migration caller filename is unavailable")
	}
	id := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	registry.withMutable(func() {
		registry.migrations = append(registry.migrations, &migrate.Migration{ID: id, Up: up, Down: down})
	})
}

// RegisterMigrationOptions 登记数据库迁移的全局执行选项。
func RegisterMigrationOptions(options ...migrate.Option) {
	registry.withMutable(func() {
		registry.migrationOptions = append(registry.migrationOptions, options...)
	})
}

// RegisterBootstrap 登记数据库迁移后的业务初始化函数。
func RegisterBootstrap(functions ...BootstrapFunc) {
	registry.withMutable(func() {
		registry.bootstraps = append(registry.bootstraps, functions...)
	})
}

// RegisterGRPC 登记业务 gRPC 服务。
func RegisterGRPC(functions ...standard.GRPCRegisterFunc) {
	registry.withMutable(func() {
		registry.grpcRegisters = append(registry.grpcRegisters, functions...)
	})
}

// RegisterGateway 登记由 google.api.http 生成的 Gateway 路由。
func RegisterGateway(functions ...standard.GatewayRegisterFunc) {
	registry.withMutable(func() {
		registry.gatewayRegisters = append(registry.gatewayRegisters, functions...)
	})
}

// RegisterRoute 登记不适合 Proto JSON 映射的额外 HTTP 接口。
func RegisterRoute(method, path string, handler standard.HandlerFunc) {
	registry.withMutable(func() {
		registry.routes = append(registry.routes, route{method: method, path: path, handler: handler})
	})
}

// RegisterServerOptions 登记不依赖运行时数据库的 Server Option。
func RegisterServerOptions(options ...standard.Option) {
	registry.withMutable(func() {
		registry.serverOptions = append(registry.serverOptions, options...)
	})
}

// RegisterServerOptionFactories 登记需要在数据库初始化后构造的 Server Option。
func RegisterServerOptionFactories(factories ...ServerOptionFactory) {
	registry.withMutable(func() {
		registry.serverOptionFactories = append(registry.serverOptionFactories, factories...)
	})
}

func (r *registrations) withMutable(update func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.frozen {
		osx.Panic("app: registrations are frozen after Run")
	}
	update()
}

func (r *registrations) freeze() registrationSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.frozen {
		osx.Panic("app: Run can only be called once")
	}
	r.frozen = true
	return registrationSnapshot{
		migrations:            slices.Clone(r.migrations),
		migrationOptions:      slices.Clone(r.migrationOptions),
		bootstraps:            slices.Clone(r.bootstraps),
		grpcRegisters:         slices.Clone(r.grpcRegisters),
		gatewayRegisters:      slices.Clone(r.gatewayRegisters),
		routes:                slices.Clone(r.routes),
		serverOptions:         slices.Clone(r.serverOptions),
		serverOptionFactories: slices.Clone(r.serverOptionFactories),
	}
}
