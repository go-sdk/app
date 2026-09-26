package app

import (
	"context"

	"github.com/go-sdk/app/internal/appstate"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/database/rdx"
)

// Redis 返回由 Run 初始化的进程级 Redis 客户端。
// 仅当 redis.enabled 为 true 时可用，且不得在包初始化期间访问。
func Redis() rdx.Client {
	client := appstate.Redis()
	if client == nil {
		osx.Panic("app: redis is not initialized")
	}
	return client
}

func openRedis(value settings) error {
	if !value.Redis.Enabled {
		return nil
	}
	client, err := rdx.Open(context.Background(), value.Redis.Config)
	if err != nil {
		return err
	}
	appstate.SetRedis(client)
	return nil
}
