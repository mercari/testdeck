package testdeck

import (
	"fmt"
	"testing"
	"time"

	"github.com/mercari/testdeck/constants"
	"github.com/mercari/testdeck/deferrer"
	"github.com/mercari/testdeck/runner"
)

/*
harness.go: A wrapper around Golang's testing.T because it has a private method that prevents us from implementing it directly

This file is separated into the following sections:

TESTDECK CODE
	Test case + lifecycle
	Statistics
CODE FROM GOLANG TESTING LIBRARY
*/

// -----
// TESTDECK CODE - Test case + lifecycle
// -----

// TestConfig is for passing special configurations for the test case
type TestConfig struct {
	// tests run in parallel by default but you can force it to run in sequential by using ParallelOff = true
	ParallelOff bool
}

// TD contains a testdeck test case + statistics to save to the DB later
// It allows us to capture functionality from testing.T
type TD struct {
	T                TestingT // wrapper on testing.T
	fatal            bool
	currentLifecycle string
	statuses         []constants.Status // stack of statuses; statuses are emitted by Error/Fatal operation or when the lifecycle completes successfully
	timings          map[string]constants.Timing
	actualName       string // name of testdeck test case (to pass to testing.T)
}

// An interface for testdeck test cases; it is implemented by the TestCase struct below
type TestCaseDelegate interface {
	ArrangeMethod(t *TD)
	ActMethod(t *TD)
	AssertMethod(t *TD)
	AfterMethod(t *TD)
}

// A struct that represents a testdeck test case
type TestCase struct {
	Arrange                  func(t *TD) // setup stage before the test
	Act                      func(t *TD) // the code you actually want to test
	Assert                   func(t *TD) // the outcomes you want to verify
	After                    func(t *TD) // clean-up steps
	deferrer.DefaultDeferrer             // deferred steps that you want to run after clean-up
}

// Interface methods
func (tc *TestCase) ArrangeMethod(t *TD) {
	timedRun(tc.Arrange, t, constants.LifecycleArrange)
}

func (tc *TestCase) ActMethod(t *TD) {
	timedRun(tc.Act, t, constants.LifecycleAct)
}

func (tc *TestCase) AssertMethod(t *TD) {
	timedRun(tc.Assert, t, constants.LifecycleAssert)
}

func (tc *TestCase) AfterMethod(t *TD) {
	timedRun(tc.After, t, constants.LifecycleAfter)
}

// This method starts the test
// t is the interface for testing.T
// tc is the interface for testdeck test cases
// options is an optional parameter for passing in special test configurations
func Test(t TestingT, tc TestCaseDelegate, options ...TestConfig) *TD {
	// FIXME: currently tests cannot be run by matching name
	tagged, matched, actualName := runner.MatchTag(t.Name())

	// start timer
	start := time.Now()
	if runner.Initialized() {
		r := runner.Instance(nil)
		r.LogEvent(fmt.Sprintf("Instantiating: %s", actualName))
	}

	// initiate testdeck test case
	td := &TD{
		T:                t,
		fatal:            false,
		currentLifecycle: constants.LifecycleTestSetup, // start in the test setup step
		timings:          make(map[string]constants.Timing),
	}

	// if test configurations struct was passed, config the settings
	if len(options) > 0 {
		if options[0].ParallelOff == false {
			td.T.Parallel()
		}
	} else {
		// if no configs were passed, turn on parallel by default
		td.T.Parallel()
	}

	// FIXME: currently tests cannot be run by matching name
	if tagged {
		if !matched {
			if runner.Initialized() {
				r := runner.Instance(nil)
				r.LogEvent("(match workaround) test not in tagged set; skipping")
			}
			return td
		}
		td.actualName = actualName
	}

	arrangeComplete := false

	// runs at the end of the test
	defer func() {
		end := time.Now()

		// clean up and set test to finished
		if !td.Skipped() || arrangeComplete {
			tc.AfterMethod(td)
		}
		td.currentLifecycle = constants.LifecycleTestFinished

		// add the final status so it is clear the test finished
		if len(td.statuses) == 0 {
			// no failure statuses, set passed
			td.setPassed()
		} else {
			// failure statuses, set failed
			td.setFailed(td.fatal)
		}

		// run deferred functions
		if d, ok := tc.(deferrer.Deferrer); ok {
			d.RunDeferred()
		}

		// save statistics to DB
		if runner.Initialized() {
			r := runner.Instance(nil)
			stats := td.makeStatistics(start, end)

			r.AddStatistics(stats)
		}
	}()
	tc.ArrangeMethod(td)
	arrangeComplete = true
	tc.ActMethod(td)
	tc.AssertMethod(td)
	return td
}

// -----
// Statistics
// -----

// Create a statistics struct for use in saving to DB later
func (c *TD) makeStatistics(start time.Time, end time.Time) *constants.Statistics {
	return &constants.Statistics{
		Name:     c.Name(),
		Failed:   c.Failed(),
		Fatal:    c.fatal,
		Statuses: c.statuses,
		Timings:  c.timings,
		Start:    start,
		End:      end,
		Duration: end.Sub(start),
	}
}

// Add result of PASSED lifecycle stage to stack
func (c *TD) setPassed() {
	status := constants.Status{
		Status:    constants.StatusPass,
		Lifecycle: c.currentLifecycle,
		Fatal:     false,
	}
	c.statuses = append(c.statuses, status)
}

// Add result of FAILED lifecycle stage to stack
func (c *TD) setFailed(fatal bool) {
	status := constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: c.currentLifecycle,
		Fatal:     fatal,
	}
	c.statuses = append(c.statuses, status)
}

// Add result of SKIPPED lifecycle stage to stack
func (c *TD) setSkipped() {
	status := constants.Status{
		Status:    constants.StatusSkip,
		Lifecycle: c.currentLifecycle,
	}
	c.statuses = append(c.statuses, status)
}

// timedRun executes fn and saves the lifecycle timing to the test case
// fn is the function to run
// t is the current test case
// lifecycle is the current test case step to save timing for
func timedRun(fn func(t *TD), t *TD, lifecycle string) {
	t.currentLifecycle = lifecycle

	timing := constants.Timing{
		Lifecycle: lifecycle,
	}

	timing.Start = time.Now()
	if fn != nil {
		timing.Started = true
		fn(t) // FIXME what if fn has a goexit (following code needs to be in defer)
		timing.Ended = true
	}
	timing.End = time.Now()
	timing.Duration = timing.End.Sub(timing.Start)

	t.timings[timing.Lifecycle] = timing
}

// -----
// CODE FROM THE GOLANG TESTING LIBRARY
// -----

// methods from testing.T
type TestingT interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	Helper()
	Parallel()
}

// Failed passes through to testing.T.Failed
func (c *TD) Failed() bool {
	return c.T.Failed()
}

// Log passes through to testing.T.Log
func (c *TD) Log(args ...interface{}) {
	c.T.Log(args...)
}

// Logf passes through to testing.T.Logf
func (c *TD) Logf(format string, args ...interface{}) {
	c.T.Logf(format, args...)
}

// Name passes through to testing.T.Name
func (c *TD) Name() string {
	// temporary workaround
	if c.actualName != "" {
		return c.actualName
	}
	return c.T.Name()
}

// Helper passes through to testing.T.Helper
func (c *TD) Helper() {
	c.T.Helper()
}

// Skipped passes through to testing.T.Skipped
func (c *TD) Skipped() bool {
	return c.T.Skipped()
}

// Fail passes through to testing.T.Fail
func (c *TD) Fail() {
	c.setFailed(false)
	c.T.Fail()
}

// Error passes through to testing.T.Error
func (c *TD) Error(args ...interface{}) {
	c.T.Helper()
	c.setFailed(false)
	c.T.Error(args...)
}

// Errorf passes through to testing.T.Errorf
func (c *TD) Errorf(format string, args ...interface{}) {
	c.T.Helper()
	c.setFailed(false)
	c.T.Errorf(format, args...)
}

// Fatal passes through to testing.T.Fatal
func (c *TD) Fatal(args ...interface{}) {
	c.T.Helper()
	c.setFailed(true)
	c.fatal = true
	c.T.Fatal(args...)
}

// Fatalf passes through to testing.T.Fatalf
func (c *TD) Fatalf(format string, args ...interface{}) {
	c.T.Helper()
	c.setFailed(true)
	c.fatal = true
	c.T.Fatalf(format, args...)
}

// Skip passes through to testing.T.Skip
func (c *TD) Skip(args ...interface{}) {
	c.setSkipped()
	c.T.Skip(args...)
}

// Skipf passes through to testing.T.Skipf
func (c *TD) Skipf(format string, args ...interface{}) {
	c.setSkipped()
	c.T.Skipf(format, args...)
}

// SkipNow passes through to testing.T.SkipNow
func (c *TD) SkipNow() {
	c.setSkipped()
	c.T.SkipNow()
}

// FailNow passes through to testing.T.FailNow
func (c *TD) FailNow() {
	c.setFailed(true)
	c.fatal = true
	c.T.FailNow()
}

// Parallel passes through to testing.T.Parallel
func (c *TD) Parallel() {
	c.T.Parallel()
}

// Run passes through to testing.T.Run
func (tc *TestCase) Run(t *testing.T, name string) {
	// this method is just a wrapper, some tests might run Test() directly so you should not do anything else here!
	// any extra actions you want to do should be added to Test() instead because that method is run every time
	t.Run(name, func(t *testing.T) {
		// Redirect to Test() to execute testdeck test case
		Test(t, tc)
	})
}
