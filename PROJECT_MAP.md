# 项目地图

## 根目录

| 路径                           | 说明                                                               |
|--------------------------------|--------------------------------------------------------------------|
| `app.go`                       | 应用运行入口、全局 Server 和生命周期编排                           |
| `config.go`                    | `core/config` 到应用运行参数的映射与校验                           |
| `database.go`                  | 默认数据库初始化和全局访问                                         |
| `redis.go`                     | 可选 Redis 初始化和全局访问                                        |
| `registry.go`                  | 迁移、初始化、任务、服务和运行期 Option 注册表                     |
| `task.go`                      | 按需初始化和访问进程级 Task Manager                                |
| `internal/appstate/`           | 生产运行时与测试 helper 共享的进程级数据库和 Redis 状态            |
| `testapp/`                     | 测试数据库初始化、Model 迁移和自动清理                             |
| `README.md`                    | 公开 API、配置约定和接入示例                                       |
| `Makefile`                     | 依赖整理、构建、静态检查和测试入口                                 |
| `.github/workflows/golang.yml` | Go 模块持续集成与标签发布流程                                      |

## 初始化顺序

1. 业务入口按需导入数据库驱动，业务包通过 `init` 调用 `Register*` 登记声明。
2. `Run` 从 `core/config` 读取统一配置并初始化日志。
3. `Run` 打开数据库并设置进程级 `DB`。
4. 启用 `redis.enabled` 时连接 Redis、执行 `PING` 并设置进程级 `Redis`。
5. `lifex.Init` 依次执行数据库迁移和业务初始化。
6. 存在已登记任务时，创建 Task Manager 并启动全部任务调度。
7. `standard.Server` 注册 gRPC、Gateway 和额外 HTTP 路由后开始监听。
8. `lifex.Wait` 等待退出信号，并按逆序停止 Task Manager、Server、Redis、数据库及日志。

## 依赖边界

```text
业务项目
  └── go-sdk/app
        ├── go-sdk/core
        ├── go-sdk/database
        ├── go-sdk/server
        └── go-sdk/taskkit
```

`app` 是有意保持强约定的集成层。底层 SDK 继续独立演进，业务项目只负责 Model、
Service、Route、Proto 和迁移等领域实现，不再重复复制基础设施初始化代码。

Model、Service 和额外 HTTP Handler 测试可以先使用 `testapp.NewDB` 按调用方指定的驱动和
DSN 初始化测试数据库，再分别直接使用 `standard/testserver.New` 或 `NewHTTP` 验证对应调用链。
