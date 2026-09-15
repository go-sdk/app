package app

import (
	"github.com/go-sdk/app/internal/appstate"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/database/dbx"
)

// DB 返回由 Run 初始化的进程级数据库实例。
// 业务 Model 可直接调用 DB().WithContext(ctx)，但不得在包初始化期间访问。
func DB() *dbx.DB {
	db := appstate.DB()
	if db == nil {
		osx.Panic("app: database is not initialized")
	}
	return db
}

func openDatabase(value settings) (*dbx.DB, error) {
	db, err := dbx.Open(value.Database.Driver, value.Database.DSN)
	if err != nil {
		return nil, err
	}
	appstate.SetDB(db)
	return db, nil
}
