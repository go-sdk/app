// Package testapp 提供使用 app 全局访问方式的测试数据库环境。
package testapp

import (
	"testing"

	"github.com/go-sdk/app/internal/appstate"
	"github.com/go-sdk/database/dbx"
)

// NewDB 使用调用方注册的驱动打开测试数据库、迁移传入的 Model，并在测试结束时自动释放资源。
// 使用 app.DB 的测试依赖进程级数据库状态，不得并行运行。
func NewDB(t testing.TB, driver, dsn string, models ...any) *dbx.DB {
	t.Helper()
	db, err := dbx.Open(driver, dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test sql database: %v", err)
	}
	installed := false
	t.Cleanup(func() {
		if installed && !appstate.CompareAndSwapDB(db, nil) {
			t.Error("test database is no longer the current app database")
		}
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	if len(models) > 0 {
		if err = db.AutoMigrate(models...); err != nil {
			t.Fatalf("migrate test database: %v", err)
		}
	}
	if !appstate.CompareAndSwapDB(nil, db) {
		t.Fatal("app database is already initialized; app database tests must not run in parallel")
	}
	installed = true
	return db
}
