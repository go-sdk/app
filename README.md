# app

`github.com/go-sdk/app` 是 `core`、`database` 和 `server` 的约定式集成层。它统一读取
默认配置、初始化数据库和可选 Redis、执行迁移和业务初始化、创建 gRPC/Gateway Server，并通过
`lifex` 管理完整生命周期。

这个模块有意保持强耦合和较少配置，适合采用同一套技术栈的新项目。需要单独替换配置、
数据库或 Server 生命周期时，应直接组合底层 SDK，而不是在 `app` 中增加兼容层。

## 安装

```shell
go get github.com/go-sdk/app
```

## 默认配置

`app` 直接使用 `core/config` 的默认实例，因此同时支持可执行文件旁的 YAML/JSON 配置、
`CONFIG_PATH` 和 `APP__` 环境变量覆盖。

运行参数通过一个内部配置结构及 `config.DecodeTo` 一次性加载。新增 app 级配置时只需扩展
该结构、默认值和校验，不需要为每个字段维护独立的 `Get` 调用。

| 配置键              | 默认值           | 说明                             |
|---------------------|------------------|----------------------------------|
| `app.name`          | 当前可执行文件名 | 日志和 Server 的服务标识         |
| `database.driver`   | `postgres`       | `mysql`、`postgres` 或 `sqlite`  |
| `database.dsn`      | 无               | 必填的原生驱动 DSN，不会写入日志 |
| `redis.enabled`     | `false`          | 是否初始化进程级 Redis 客户端    |
| `redis.mode`        | `standalone`     | `standalone` 或 `sentinel`       |
| `redis.addresses`   | `127.0.0.1:6379` | Redis 或 Sentinel 地址列表       |
| `redis.master_name` | 空               | Sentinel 模式的 master 名称      |
| `redis.database`    | `0`              | Redis 数据库编号                 |
| `server.address`    | `:8080`          | gRPC 与 HTTP 共用的监听地址      |
| `server.reflection` | `false`          | 是否启用 gRPC Reflection         |
| `auth.jwt_secret`   | 空               | 非空时启用标准 JWT 鉴权          |
| `log.file.path`     | 空               | 非空时同时写入滚动日志文件       |

连接池继续使用 `database` 模块已有的 `database.pool.*` 或
`database.<driver>.pool.*` 配置，不在本模块重复定义。

Redis 可继续配置 `username`、`password`、`sentinel_username`、`sentinel_password`、
`client_name`、连接超时、连接池和 TLS 参数，字段与 `database/rdx.Config` 一致。配置中的地址和
凭据不会写入初始化日志。启用后可通过 `app.Redis()` 获取客户端；初始化会先执行 `PING`，失败时
阻止 Server 启动。Sentinel 锁仍需配合业务数据库 fencing token 使用。

## 最小入口

```go
package main

import (
	_ "example/internal/model"
	_ "example/internal/route"
	_ "example/internal/service"

	"github.com/go-sdk/app"
	_ "github.com/go-sdk/database/dbx/postgres"
)

func main() {
	app.Main()
}
```

业务包的 `init` 只登记声明。数据库连接、迁移和端口监听均在 `app.Run` 中发生，导入
Model 包不会产生数据库或网络副作用。

数据库驱动必须由具体项目按需空导入，并与 `database.driver` 保持一致。上例只导入
PostgreSQL；MySQL 和 SQLite 项目分别改为 `dbx/mysql` 和 `dbx/sqlite`。`app` 不会隐式
导入全部驱动。

## Model 与迁移

建议每个 Model 使用一个文件，并把该 Model 的查询和写入方法放在同一文件中。方法直接
使用全局数据库，不需要每个 Service 重复注入或保存 `*dbx.DB`：

```go
package model

import (
	"context"

	"github.com/go-sdk/app"
)

type User struct {
	ID   int64
	Name string
}

func FindUser(ctx context.Context, id int64) (*User, error) {
	var value User
	err := app.DB().WithContext(ctx).First(&value, "id = ?", id).Error
	return &value, err
}
```

迁移文件名必须符合 `YYYYMMDD_HHMMSS_NN_description.go`。迁移通过文件级注册直接纳入
全局执行顺序，不再需要项目级迁移集合：

```go
func init() {
	app.RegisterMigration(up, nil)
}

func up(db *dbx.DB) error {
	return db.AutoMigrate(&User{})
}
```

需要在迁移后创建内置数据时，通过 `RegisterBootstrap` 注册。初始化失败会阻止 Server
启动。

## gRPC 与 Gateway

普通业务接口仍以 Proto 为唯一协议来源。Service 包同时注册 gRPC 实现和生成的 Gateway
绑定：

```go
func init() {
	app.RegisterGRPC(func(server grpc.ServiceRegistrar) {
		v1.RegisterUserServiceServer(server, NewUserService())
	})
	app.RegisterGateway(v1.RegisterUserServiceHandlerFromEndpoint)
}
```

需要数据库构造鉴权拦截器等 Server Option 时，使用
`RegisterServerOptionFactories`；工厂在数据库初始化后、Server 创建前执行。静态 Option
直接使用 `RegisterServerOptions`。

## 额外 HTTP 接口

multipart 上传、文件下载等不适合 Proto JSON 映射的接口放在业务项目的 `route` 包，通过
`RegisterRoute` 注册：

```go
func init() {
	app.RegisterRoute(http.MethodPost, "/api/v1/files", uploadFile)
}
```

普通 JSON API 不应绕过 Proto 和生成的 Gateway 绑定。

## 单元测试

使用 `app.DB()` 的 Model、Service 或额外 HTTP Handler 测试可以通过 `testapp.NewDB` 使用
调用方指定的驱动和 DSN 创建测试数据库，并迁移所需 Model：

```go
func TestCreateUser(t *testing.T) {
	testapp.NewDB(t, testDriver, testDSN, &User{}, &Role{}, &UserRole{})

	if err := CreateUser(context.Background(), &User{Username: "tester"}); err != nil {
		t.Fatal(err)
	}
}
```

具体数据库驱动必须由测试项目按需导入并注册。helper 会在测试结束时移除全局数据库并
关闭连接，但不会创建或删除测试数据库本身；使用共享测试库时应由调用方保证数据隔离。
由于 `app.DB()` 是进程级状态，使用该 helper 的测试不得调用 `t.Parallel()`。Service 的
gRPC 链路继续直接使用 `standard/testserver.New`，额外 HTTP 接口使用
`standard/testserver.NewHTTP`；app 不再重复包装 Server 测试能力，也不在单元测试中执行
已注册的生产迁移和 Bootstrap。

## 生命周期约束

- 所有 `Register*` 必须在 `Run` 前完成，`Run` 后继续注册会 panic。
- `DB`、`Redis` 和 `Server` 只在 `Run` 初始化过程中及之后可用，不得在业务包 `init` 中访问；
  `Redis` 还要求 `redis.enabled=true`。
- `Run` 是进程级单次入口，不支持停止后再次启动。
- 初始化失败会触发已登记资源的逆序清理；信号退出和主动退出由 `lifex` 统一处理。
