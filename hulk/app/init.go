package app

import "os"

var (
	RepoOwner = os.Getenv("GROWERLAB_REPO_OWNER")
	RepoPath  = os.Getenv("GROWERLAB_REPO_NAME")
)

func init() {
	ErrPanic(InitConfig())
	ErrPanic(InitRedis())

	app = &App{
		dispatcher: &EventDispatch{},
	}
	app.RegisterHook(&HookEvent{})
}
