package runner

import (
	"bytes"
	"fmt"
	"io"
	tdlog "log"
	"os"
	"reflect"
	"regexp"
	"sync"
	"testing"
	"unsafe"

	"github.com/mercari/testdeck/constants"
)

/*
runner.go: This implements a custom version of Golang testing library's Run(). It only works with Tests (Benchmark and Example are not supported)

NOTE:
Caution:
	THIS RUNNER IS POTENTIALLY UNSAFE/FRAGILE. Consider it experimental.
Why:
 Go standard lib testing package does not export a lot of functionality making
 it hard to work with safely (requires us to use `reflect` and `unsafe` to
 access hidden functionality). This also means we are accessing implementation
 details that can change between Go versions (fragile).

For now we will go with this implementation because the cost of writing our
own Runner from scratch is likely higher than depending on the implementation
of the Go standard runner. Eventually we will have to implement our own
runner when there is more time.
*/

// RealStdout will point to the real stdout
var RealStdout *os.File
var printStdout = true
var printOutputToEventLog = false
var instance Runner
var once sync.Once

// Regex for matching test cases so that single test cases can be run
// FIXME: This feature is currently not working (single test cases cannot be run by name)
var reMatchTag = regexp.MustCompile("^(.*)\x00(.*)$")
var reFilterTags = regexp.MustCompile("\\^.+\x00")
var EnableMatchWorkaround = true

// EventLogger will log test events
type EventLogger interface {
	Log(message string)
}

// Contains the custom test runner and test run constants (output, logs, etc.)
type runner struct {
	m            TestRunner
	deps         *TestDeps
	output       string
	stats        []constants.Statistics
	eventLogger  EventLogger
	matchRe      *regexp.Regexp
	matchPattern string
}

// Interface for the custom test runner (contains Golang's Run() and some other custom methods that we need for recording statistics, etc.)
type Runner interface {
	Run()
	AddStatistics(stats *constants.Statistics)
	Statistics() []constants.Statistics
	ClearStatistics()
	Match(pattern string) error
	PrintToStdout(yes bool)
	PrintOutputToEventLog(yes bool)
	SetEventLogger(e EventLogger)
	LogEvent(message string)
	ReportStatistics()
	Passed() bool
	Output() string
}

// This is a custom version of Golang testing's type M (a test runner struct)
type TestRunner interface {
	Run() int
}

// -----
// TEST RUNNER
// -----

// This is pulled out so we can replace it for unit testing. The Go testing
// package has too much assumed global state so we can't actually use the real
// thing for unit tests.
var runnerMainStart = func(deps *TestDeps, tests []testing.InternalTest) {
	// We need to instantiate our own "m" so we can feed it our implementation of
	// testDeps. This allows us to control the running match pattern between Runs.
	m2 := testing.MainStart(deps, tests, make([]testing.InternalBenchmark, 0), make([]testing.InternalExample, 0))
	m2.Run()
}

// GetInstance returns the runner instance. Only the first invocation of this
// method will set the "m". This should not matter because the testing framework
// currently does not let us safely create our own "m" so only one instance
// should always exist.
//
// Internally the testdeck package will invoke this method with 'nil' in order
// to access the runner.
func Instance(m TestRunner) Runner {
	if m == nil && instance == nil {
		panic("Accessing an uninitialized Runner instance. You probably never gave a valid 'm' parameter.")
	}
	once.Do(func() {
		instance = newInstance(m)
	})
	return instance
}

// make accessible for testing
func newInstance(m TestRunner) Runner {
	return &runner{
		m:    m,
		deps: &TestDeps{},
	}
}

// Initialized returns true if the instance is initialized.
func Initialized() bool {
	return instance != nil
}

// Run starts the test runner
func (r *runner) Run() {

	// FIXME: Filtering tests to run by name is not working right now
	tests := filterTestsWorkaround(r.matchRe, getInternalTests(r.m), EnableMatchWorkaround, r.matchPattern)

	// Create a tee to duplicate stdout writes to a buffer we can read later.
	// idea from: https://stackoverflow.com/a/10476304
	RealStdout := os.Stdout
	rp, wp, _ := os.Pipe()
	outChannel := make(chan string)
	go func() {
		var buf bytes.Buffer
		if printStdout || printOutputToEventLog {
			var writers []io.Writer

			if printStdout {
				writers = append(writers, RealStdout)
			}

			if printOutputToEventLog {
				writers = append(writers, NewEventWriter(r))
			}

			teeStdout := io.TeeReader(rp, io.MultiWriter(writers...))
			_, err := io.Copy(&buf, teeStdout)
			if err != nil {
				tdlog.Println("testdeck output capture issue, io.Copy err:", err)
			}
		} else {
			_, err := io.Copy(&buf, rp)
			if err != nil {
				tdlog.Println("testdeck output capture issue, io.Copy err:", err)
			}
		}
		outChannel <- buf.String()
	}()

	os.Stdout = wp

	runnerMainStart(r.deps, tests)

	wp.Close()             // close the pipe so the io.Copy gets EOF
	os.Stdout = RealStdout // reset stdout

	r.output = <-outChannel
	rp.Close()

	// FIXME: Running individual test cases by matching name is not working now
	if EnableMatchWorkaround {
		r.output = reFilterTags.ReplaceAllString(r.output, "")
	}

	// FIXME: Each test case is saving the entire test run's output. This should be fixed so that only the test case's output is saved.
	for i, _ := range r.stats {
		r.stats[i].Output = r.output
	}
}

// -----
// STATISTICS, OUTPUT, AND LOGGING
// -----

func (r *runner) AddStatistics(stats *constants.Statistics) {
	r.stats = append(r.stats, *stats)
}

func (r *runner) Statistics() []constants.Statistics {
	return r.stats
}

// ClearStatistics resets the stats to nothing.
func (r *runner) ClearStatistics() {
	r.stats = make([]constants.Statistics, 0)
}

func (r *runner) ReportStatistics() {
	for i, s := range r.stats {
		fmt.Println(i, s.Failed, s.Name)
	}
}

func (r *runner) Passed() bool {
	for _, s := range r.stats {
		if s.Failed {
			return false
		}
	}
	return true
}

func (r *runner) Output() string {
	return r.output
}

func (r *runner) PrintToStdout(yes bool) {
	printStdout = yes
}

func (r *runner) PrintOutputToEventLog(yes bool) {
	printOutputToEventLog = yes
}

func (r *runner) SetEventLogger(e EventLogger) {
	r.eventLogger = e
}

func (r *runner) LogEvent(message string) {
	if r.eventLogger != nil {
		r.eventLogger.Log(message)
	}
}

// -----
// TEST NAME MATCHING
// FIXME: This feature is not working now, tests cannot be run individually by name
// -----

// MatchTag returns ok = true if the name has a test tag. If the tag is present,
// then "match" will indicate if the test name matched the tagged pattern. If
// there is a test tag, the returned name will be the actual test name minus the
// tag.
//
// If there is no tag the function returns the same given name and ok = false,
// matched = false.
//
// This should only be used as a workaround until we implement a real test
// runner.
func MatchTag(name string) (ok bool, matched bool, actualName string) {
	if reMatchTag.MatchString(name) {
		parts := reMatchTag.FindStringSubmatch(name)
		tagPattern := parts[1]
		actual := parts[2]
		re := regexp.MustCompile(tagPattern)
		return true, re.MatchString(actual), actual
	}

	return false, false, name
}

// Match sets the regular expression pattern to filter tests to run.
func (r *runner) Match(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	r.matchRe = re
	r.matchPattern = pattern // for temporary workaround
	return nil
}

// Temporary workaround to run individual test cases by name
// FIXME: This is not working now, individual test cases cannot be run by name
func filterTestsWorkaround(re *regexp.Regexp, tests []testing.InternalTest, matchWorkaround bool, rePattern string) []testing.InternalTest {
	if matchWorkaround && rePattern != ".*" {
		var tagged []testing.InternalTest

		for _, test := range tests {
			clone := test
			clone.Name = rePattern + "\x00" + test.Name
			tagged = append(tagged, clone)
		}

		return tagged
	}

	return filterTests(re, tests)
}

// FIXME: This is not working now, individual test cases cannot be run by name
func filterTests(re *regexp.Regexp, tests []testing.InternalTest) []testing.InternalTest {
	var filtered []testing.InternalTest

	for _, test := range tests {
		if re.MatchString(test.Name) {
			filtered = append(filtered, test)
		}
	}

	return filtered
}

// -----
// CODE COPIED FROM GOLANG TESTING LIBRARY
// -----

func getInternalTests(m TestRunner) []testing.InternalTest {
	internalTestsIndex, found := getInternalTestsFieldIndex(m)
	if !found {
		panic("Could not find []InternalTest via reflect. Perhaps you updated the Go library version?")
	}

	// https://stackoverflow.com/a/43918797
	// Access an unexported field by index
	rs := reflect.ValueOf(m).Elem()
	rf := rs.Field(internalTestsIndex)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	internalTests := make([]testing.InternalTest, rf.Len())
	for i := 0; i < rf.Len(); i++ {
		val := rf.Index(i)
		iTest := val.Interface().(testing.InternalTest)
		internalTests[i] = iTest
	}

	return internalTests
}

// getInternalTestsFieldIndex returns the index of the []InternalTest slice if
// it exists. If it does not exist the "ok" value will be set to false.
func getInternalTestsFieldIndex(m TestRunner) (index int, ok bool) {
	rs := reflect.ValueOf(m).Elem()
	for i := 0; i < rs.NumField(); i++ {
		field := rs.Field(i)
		// filter only slice kinds
		if reflect.Slice == field.Kind() {
			it := (*testing.InternalTest)(nil)
			// match only testing.InternalTest type
			if reflect.TypeOf(it).Elem() == field.Type().Elem() {
				return i, true
			}
		}
	}
	return 0, false
}
