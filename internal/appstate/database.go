package appstate

import (
	"sync/atomic"

	"github.com/go-sdk/database/dbx"
)

var defaultDB atomic.Pointer[dbx.DB]

// DB 返回当前进程级数据库实例，未初始化时返回 nil。
func DB() *dbx.DB {
	return defaultDB.Load()
}

// SetDB 设置生产运行时使用的进程级数据库实例。
func SetDB(db *dbx.DB) {
	defaultDB.Store(db)
}

// CompareAndSwapDB 仅供测试环境安全地安装和移除临时数据库实例。
func CompareAndSwapDB(old, new *dbx.DB) bool {
	return defaultDB.CompareAndSwap(old, new)
}
