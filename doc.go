// Package app 是 core、database 和 server 的约定式集成层，提供应用运行时。
// 它统一读取默认配置、初始化数据库、执行迁移和业务初始化、创建 gRPC/Gateway
// Server，并通过 lifex 管理完整生命周期。
//
// 最小入口只需在业务包的 init 中登记声明，然后调用 app.Main：
//
//	import (
//		"github.com/go-sdk/app"
//		_ "example/internal/model"
//		_ "example/internal/service"
//		_ "github.com/go-sdk/database/dbx/postgres"
//	)
//
//	func main() { app.Main() }
//
// 数据库驱动必须由具体项目按需空白导入，并与 database.driver 配置保持一致。
//
// 安装：
//
//	go get github.com/go-sdk/app
//
// 这个模块有意保持强耦合和较少配置，适合采用同一套技术栈的新项目。需要单独
// 替换配置、数据库或 Server 生命周期时，应直接组合底层 SDK，而不是在 app 中
// 增加兼容层。
package app
