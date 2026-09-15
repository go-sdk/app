package testapp_test

import (
	"path/filepath"
	"testing"

	"github.com/go-sdk/app"
	"github.com/go-sdk/app/testapp"
	_ "github.com/go-sdk/database/dbx/sqlite"
)

type testRecord struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func TestNewDB(t *testing.T) {
	for _, name := range []string{"first", "second"} {
		t.Run(name, func(t *testing.T) {
			testapp.NewDB(t, "sqlite", filepath.Join(t.TempDir(), "test.db"), &testRecord{})
			value := testRecord{Name: "value"}
			if err := app.DB().Create(&value).Error; err != nil {
				t.Fatal(err)
			}
			var count int64
			if err := app.DB().Model(&testRecord{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("unexpected record count: %d", count)
			}
		})
	}
}
