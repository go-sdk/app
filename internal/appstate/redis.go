package appstate

import (
	"sync/atomic"

	"github.com/go-sdk/database/rdx"
)

type redisHolder struct {
	client rdx.Client
}

var defaultRedis atomic.Pointer[redisHolder]

// Redis 返回当前进程级 Redis 客户端，未初始化时返回 nil。
func Redis() rdx.Client {
	holder := defaultRedis.Load()
	if holder == nil {
		return nil
	}
	return holder.client
}

// SetRedis 设置生产运行时使用的进程级 Redis 客户端。
func SetRedis(client rdx.Client) {
	defaultRedis.Store(&redisHolder{client: client})
}
