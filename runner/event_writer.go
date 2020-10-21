package runner

/*
event_writer.go: A custom event logger used for writing test result output
*/

type eventWriter struct {
	runner Runner
}

func NewEventWriter(runner Runner) *eventWriter {
	return &eventWriter{
		runner: runner,
	}
}

func (e *eventWriter) Write(p []byte) (n int, err error) {
	e.runner.LogEvent(string(p))
	return len(p), nil
}