package executor

type Executor interface {
	RunOnce()
	Command() string
}
