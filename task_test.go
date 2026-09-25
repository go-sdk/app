package app

import (
	"context"
	"testing"
	"time"

	"github.com/go-sdk/taskkit"
)

func TestTaskInitializerCreatesConfiguredManager(t *testing.T) {
	factoryCalled := false
	registered := registrationSnapshot{
		tasks: []registeredTask{
			{
				id:       "job-1",
				schedule: taskkit.Once(time.Now().Add(time.Hour)),
				task:     func(*taskkit.Context) error { return nil },
			},
		},
		taskOptionFactories: []TaskManagerOptionFactory{
			func() (taskkit.ManagerOption, error) {
				factoryCalled = true
				return taskkit.WithLocation(time.UTC), nil
			},
		},
	}

	if err := initializeTaskManager(registered); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := TaskManager().Shutdown(ctx); err != nil {
			t.Error(err)
		}
	})
	if !factoryCalled {
		t.Fatal("task manager option factory was not called")
	}
	if _, ok := TaskManager().Get("job-1"); !ok {
		t.Fatal("registered task was not added")
	}
}
