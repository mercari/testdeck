package deferrer

/*
deferrer.go: Helper functions for building the defer stack (to force test steps to run in order)
*/

type Deferrer interface {
	Defer(fn func())
	RunDeferred()
}

type DefaultDeferrer struct {
	deferStack []func()
}

// Pushes a function info the defer stack to run later
func (d *DefaultDeferrer) Defer(fn func()) {
	d.deferStack = append(d.deferStack, fn)
}

// Run deferred functions in LIFO order
func (d *DefaultDeferrer) RunDeferred() {
	for i := len(d.deferStack) - 1; i >= 0; i-- {
		d.deferStack[i]()
	}
}
