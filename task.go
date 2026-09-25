package app

import (
	"sync/atomic"

	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/lifex"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/taskkit"
)

var defaultTaskManager atomic.Pointer[taskkit.Manager]

// TaskManager 返回由 Run 按需初始化的进程级任务管理器。
func TaskManager() *taskkit.Manager {
	manager := defaultTaskManager.Load()
	if manager == nil {
		osx.Panic("app: task manager is not initialized")
	}
	return manager
}

func registerTaskInitializer(registered registrationSnapshot) {
	if len(registered.tasks) == 0 {
		return
	}
	lifex.OnInit(func() error { return initializeTaskManager(registered) })
}

func initializeTaskManager(registered registrationSnapshot) error {
	options := make([]taskkit.ManagerOption, 0, len(registered.taskOptionFactories))
	for _, factory := range registered.taskOptionFactories {
		if factory == nil {
			return errx.New("task manager option factory must not be nil")
		}
		option, err := factory()
		if err != nil {
			return errx.Wrap(err, "create task manager option")
		}
		if option == nil {
			return errx.New("task manager option factory returned nil")
		}
		options = append(options, option)
	}
	manager, err := taskkit.NewManager(options...)
	if err != nil {
		return errx.Wrap(err, "initialize task manager")
	}
	defaultTaskManager.Store(manager)
	for _, item := range registered.tasks {
		if _, err = manager.Add(item.id, item.schedule, item.task, item.options...); err != nil {
			return errx.Wrapf(err, "register task %q", item.id)
		}
	}
	return nil
}
