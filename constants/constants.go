package constants

import "time"

/*
constants.go: Contains constants and struct definitions
*/

// Test case status
const (
	StatusFail = "Fail"
	StatusPass = "Pass"
	StatusSkip = "Skip"
)

// Test case stages
const (
	LifecycleTestSetup    = "FrameworkTestSetup"
	LifecycleArrange      = "Arrange"
	LifecycleAct          = "Act"
	LifecycleAssert       = "Assert"
	LifecycleAfter        = "After"
	LifecycleTestFinished = "FrameworkTestFinished"
)

// Status stores the lifecycle stage and test status
type Status struct {
	Status    string
	Lifecycle string
	Fatal     bool
}

// Timing stores the start, end, and duration of a lifecycle stage
type Timing struct {
	Lifecycle string
	Start     time.Time
	End       time.Time
	Duration  time.Duration
	Started   bool
	Ended     bool
}

// Statistics are the test results that will be saved to the DB
type Statistics struct {
	Name     string
	Failed   bool
	Fatal    bool
	Statuses []Status
	Timings  map[string]Timing
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Output   string
}

const DefaultHttpTimeout = time.Second * 30 // default HTTP client timeout

// test result constants
const (
	ResultPass = "PASS"
	ResultFail = "FAIL"
)
