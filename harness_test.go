package testdeck

import (
	"errors"
	"sync"
	"testing"

	"github.com/mercari/testdeck/constants"
	. "github.com/mercari/testdeck/fname"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit Tests for testdeck

// mock testing.TB implementation

func newMockT() *mockT {
	return &mockT{
		callCount: &incMap{},
		argsList:  make(map[string][][]interface{}),
		retVal: map[string]interface{}{
			"Skipped": false,
			"Failed":  false,
			"Name":    "",
		},
	}
}

type incMap struct {
	sync.Map
}

func (m *incMap) increment(key interface{}) {
	value := m.get(key)
	m.Store(key, value+1)
}

func (m *incMap) get(key interface{}) int {
	existing, ok := m.Load(key)
	var value int
	if ok {
		value = existing.(int)
	}
	return value
}

type mockT struct {
	testing.T
	callCount *incMap
	argsList  map[string][][]interface{}
	retVal    map[string]interface{}

	ErrorfFormat []string
	FatalfFormat []string
	SkipfFormat  []string
}

func (t *mockT) returnValue(funcName string, value interface{}) {
	t.retVal[funcName] = value
}

func (t *mockT) logCall() {
	t.callCount.increment(Fname(1))
}

func (t *mockT) logCallArgs(args ...interface{}) {
	fn := Fname(1)
	t.callCount.increment(Fname(1))
	t.argsList[fn] = append(t.argsList[fn], args)
}

func (t *mockT) Error(args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Errorf(format string, args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Fatal(args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Fatalf(format string, args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Skip(args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Skipf(format string, args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Log(args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) Logf(format string, args ...interface{}) {
	t.logCallArgs(args...)
}

func (t *mockT) SkipNow() {
	t.logCall()
}

func (t *mockT) FailNow() {
	t.logCall()
}

func (t *mockT) Fail() {
	t.logCall()
}

func (t *mockT) Helper() {
	t.logCall()
}

func (t *mockT) Failed() bool {
	fn := Fname()
	t.callCount.increment(fn)
	return t.retVal[fn].(bool)
}

func (t *mockT) Skipped() bool {
	fn := Fname()
	t.callCount.increment(fn)
	return t.retVal[fn].(bool)
}

func (t *mockT) Name() string {
	fn := Fname()
	t.callCount.increment(fn)
	return t.retVal[fn].(string)
}

// Unit Tests

func Test_TD_ArgMethodsPassThrough(t *testing.T) {
	// Arrange
	mock := newMockT()
	td := TD{
		T: mock,
	}
	err := errors.New("Error")

	// Act
	td.Error(err)
	td.Fatal(err)
	td.Skip(err)
	td.Log(err)

	// Assert
	assert.Equal(t, 1, mock.callCount.get("Error"))
	assert.Equal(t, 0, mock.callCount.get("Errorf"))
	assert.Equal(t, 1, mock.callCount.get("Fatal"))
	assert.Equal(t, 0, mock.callCount.get("Fatalf"))
	assert.Equal(t, 1, mock.callCount.get("Skip"))
	assert.Equal(t, 0, mock.callCount.get("Skipf"))
	assert.Equal(t, 1, mock.callCount.get("Log"))
	assert.Equal(t, 0, mock.callCount.get("Logf"))
	assert.Equal(t, 0, mock.callCount.get("SkipNow"))
	assert.Equal(t, 0, mock.callCount.get("FailNow"))
	assert.Equal(t, 0, mock.callCount.get("Fail"))
	require.Equal(t, 1, len(mock.argsList["Error"]))
	require.Equal(t, 1, len(mock.argsList["Fatal"]))
	require.Equal(t, 1, len(mock.argsList["Skip"]))
	require.Equal(t, 1, len(mock.argsList["Log"]))
	require.Equal(t, 1, len(mock.argsList["Error"][0]))
	require.Equal(t, 1, len(mock.argsList["Fatal"][0]))
	require.Equal(t, 1, len(mock.argsList["Skip"][0]))
	require.Equal(t, 1, len(mock.argsList["Log"][0]))
	assert.Equal(t, err, mock.argsList["Error"][0][0])
	assert.Equal(t, err, mock.argsList["Fatal"][0][0])
	assert.Equal(t, err, mock.argsList["Skip"][0][0])
	assert.Equal(t, err, mock.argsList["Log"][0][0])
	// assert.Equal(t, 0, mock.callCount.get("Helper")) // methods may arbitrarily call Helper
}

func Test_TD_ArgfMethodsPassThrough(t *testing.T) {
	// Arrange
	mock := newMockT()
	td := TD{
		T: mock,
	}
	err := errors.New("Error")
	format := "format string"

	// Act
	td.Errorf(format, err)
	td.Fatalf(format, err)
	td.Skipf(format, err)
	td.Logf(format, err)

	// Assert
	assert.Equal(t, 0, mock.callCount.get("Error"))
	assert.Equal(t, 1, mock.callCount.get("Errorf"))
	assert.Equal(t, 0, mock.callCount.get("Fatal"))
	assert.Equal(t, 1, mock.callCount.get("Fatalf"))
	assert.Equal(t, 0, mock.callCount.get("Skip"))
	assert.Equal(t, 1, mock.callCount.get("Skipf"))
	assert.Equal(t, 0, mock.callCount.get("Log"))
	assert.Equal(t, 1, mock.callCount.get("Logf"))
	assert.Equal(t, 0, mock.callCount.get("SkipNow"))
	assert.Equal(t, 0, mock.callCount.get("FailNow"))
	assert.Equal(t, 0, mock.callCount.get("Fail"))
	require.Equal(t, 1, len(mock.argsList["Errorf"]))
	require.Equal(t, 1, len(mock.argsList["Fatalf"]))
	require.Equal(t, 1, len(mock.argsList["Skipf"]))
	require.Equal(t, 1, len(mock.argsList["Logf"]))
	require.Equal(t, 1, len(mock.argsList["Errorf"][0]))
	require.Equal(t, 1, len(mock.argsList["Fatalf"][0]))
	require.Equal(t, 1, len(mock.argsList["Skipf"][0]))
	require.Equal(t, 1, len(mock.argsList["Logf"][0]))
	assert.Equal(t, err, mock.argsList["Errorf"][0][0])
	assert.Equal(t, err, mock.argsList["Fatalf"][0][0])
	assert.Equal(t, err, mock.argsList["Skipf"][0][0])
	assert.Equal(t, err, mock.argsList["Logf"][0][0])
	// assert.Equal(t, 0, mock.callCount.get("Helper")) // methods may arbitrarily call Helper
}

func Test_TD_NoArgMethodsPassThrough(t *testing.T) {
	// Arrange
	mock := newMockT()
	td := TD{
		T: mock,
	}

	// Act
	td.SkipNow()
	td.FailNow()
	td.Fail()
	td.Helper()

	// Assert
	assert.Equal(t, 0, mock.callCount.get("Error"))
	assert.Equal(t, 0, mock.callCount.get("Errorf"))
	assert.Equal(t, 0, mock.callCount.get("Fatal"))
	assert.Equal(t, 0, mock.callCount.get("Fatalf"))
	assert.Equal(t, 0, mock.callCount.get("Skip"))
	assert.Equal(t, 0, mock.callCount.get("Skipf"))
	assert.Equal(t, 0, mock.callCount.get("Log"))
	assert.Equal(t, 0, mock.callCount.get("Logf"))
	assert.Equal(t, 1, mock.callCount.get("SkipNow"))
	assert.Equal(t, 1, mock.callCount.get("FailNow"))
	assert.Equal(t, 1, mock.callCount.get("Fail"))
	assert.NotZero(t, mock.callCount.get("Helper"))
}

func Test_TD_ReturnMethodsPassThrough(t *testing.T) {
	// Arrange
	mock := newMockT()
	const (
		skippedVal = true
		failedVal  = true
		nameVal    = "test name"
	)
	mock.returnValue("Skipped", skippedVal)
	mock.returnValue("Failed", failedVal)
	mock.returnValue("Name", nameVal)
	td := TD{
		T: mock,
	}

	// Act
	returnedSkipped := td.Skipped()
	returnedFailed := td.Failed()
	returnedName := td.Name()

	// Assert
	assert.Equal(t, 0, mock.callCount.get("Error"))
	assert.Equal(t, 0, mock.callCount.get("Errorf"))
	assert.Equal(t, 0, mock.callCount.get("Fatal"))
	assert.Equal(t, 0, mock.callCount.get("Fatalf"))
	assert.Equal(t, 0, mock.callCount.get("Skip"))
	assert.Equal(t, 0, mock.callCount.get("Skipf"))
	assert.Equal(t, 0, mock.callCount.get("Log"))
	assert.Equal(t, 0, mock.callCount.get("Logf"))
	assert.Equal(t, 0, mock.callCount.get("SkipNow"))
	assert.Equal(t, 0, mock.callCount.get("FailNow"))
	assert.Equal(t, 0, mock.callCount.get("Fail"))
	// assert.Equal(t, 0, mock.callCount.get("Helper")) // methods may arbitrarily call Helper
	assert.Equal(t, 1, mock.callCount.get("Skipped"))
	assert.Equal(t, 1, mock.callCount.get("Failed"))
	assert.Equal(t, 1, mock.callCount.get("Name"))
	assert.Equal(t, skippedVal, returnedSkipped)
	assert.Equal(t, failedVal, returnedFailed)
	assert.Equal(t, nameVal, returnedName)
}

func Test_Test_ShouldExecuteAllLifecycleMethodsInOrder(t *testing.T) {
	// Arrange
	mock := newMockT()
	var callIndex = 0
	callCounts := make(map[string]int)
	callIndexes := make(map[string]int)

	testCase := &TestCase{
		Arrange: func(t *TD) {
			callIndex++
			callIndexes[constants.LifecycleArrange] = callIndex
			callCounts[constants.LifecycleArrange]++
		},
		Act: func(t *TD) {
			callIndex++
			callIndexes[constants.LifecycleAct] = callIndex
			callCounts[constants.LifecycleAct]++
		},
		Assert: func(t *TD) {
			callIndex++
			callIndexes[constants.LifecycleAssert] = callIndex
			callCounts[constants.LifecycleAssert]++
		},
		After: func(t *TD) {
			callIndex++
			callIndexes[constants.LifecycleAfter] = callIndex
			callCounts[constants.LifecycleAfter]++
		},
	}

	// Act
	td := Test(mock, testCase, TestConfig{ParallelOff: true})

	// Assert
	assert.Equal(t, 1, callCounts[constants.LifecycleArrange])
	assert.Equal(t, 1, callCounts[constants.LifecycleAct])
	assert.Equal(t, 1, callCounts[constants.LifecycleAssert])
	assert.Equal(t, 1, callCounts[constants.LifecycleAfter])

	assert.Equal(t, 1, callIndexes[constants.LifecycleArrange])
	assert.Equal(t, 2, callIndexes[constants.LifecycleAct])
	assert.Equal(t, 3, callIndexes[constants.LifecycleAssert])
	assert.Equal(t, 4, callIndexes[constants.LifecycleAfter])

	assert.Equal(t, constants.Status{
		Status:    constants.StatusPass,
		Lifecycle: constants.LifecycleTestFinished,
		Fatal:     false,
	}, td.statuses[0])
}

func Test_Test_ShouldExecuteNoNilLifecycleMethods(t *testing.T) {
	// Arrange
	mock := newMockT()
	testCase := &TestCase{}

	// Act
	Test(mock, testCase, TestConfig{ParallelOff: true})

	// Assert
	assert.Equal(t, 0, mock.callCount.get(constants.LifecycleArrange))
	assert.Equal(t, 0, mock.callCount.get(constants.LifecycleAct))
	assert.Equal(t, 0, mock.callCount.get(constants.LifecycleAssert))
	assert.Equal(t, 0, mock.callCount.get(constants.LifecycleAfter))
}

func Test_Test_ShouldAllowErrorInEachLifecycleMethod(t *testing.T) {
	// Arrange
	mock := newMockT()

	testCase := &TestCase{
		Arrange: func(t *TD) {
			t.Error()
		},
		Act: func(t *TD) {
			t.Error()
		},
		Assert: func(t *TD) {
			t.Error()
		},
		After: func(t *TD) {
			t.Error()
		},
	}

	// Act
	td := Test(mock, testCase, TestConfig{ParallelOff: true})

	// Assert
	const wantLenStatus = 5
	gotLenStatuses := len(td.statuses)
	if gotLenStatuses != wantLenStatus {
		t.Fatalf("Want %d statuses, got %d", wantLenStatus, gotLenStatuses)
	}

	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleArrange,
		Fatal:     false,
	}, td.statuses[0])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleAct,
		Fatal:     false,
	}, td.statuses[1])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleAssert,
		Fatal:     false,
	}, td.statuses[2])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleAfter,
		Fatal:     false,
	}, td.statuses[3])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleTestFinished,
		Fatal:     false,
	}, td.statuses[4])
}

func Test_Test_ShouldSetTimingsForLifecycleMethods(t *testing.T) {
	allLifecycles := []string{
		constants.LifecycleArrange,
		constants.LifecycleAct,
		constants.LifecycleAssert,
		constants.LifecycleAfter,
	}

	cases := map[string]struct {
		testCase       *TestCase
		wantLenTimings int
		wantLifecycle  string
	}{
		constants.LifecycleArrange: {
			testCase: &TestCase{
				Arrange: func(t *TD) {
					t.Error()
				},
			},
			wantLifecycle: constants.LifecycleArrange,
		},
		constants.LifecycleAct: {
			testCase: &TestCase{
				Act: func(t *TD) {
					t.Error()
				},
			},
			wantLifecycle: constants.LifecycleAct,
		},
		constants.LifecycleAssert: {
			testCase: &TestCase{
				Assert: func(t *TD) {
					t.Error()
				},
			},
			wantLifecycle: constants.LifecycleAssert,
		},
		constants.LifecycleAfter: {
			testCase: &TestCase{
				After: func(t *TD) {
					t.Error()
				},
			},
			wantLifecycle: constants.LifecycleAfter,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			mock := newMockT()

			// Act
			td := Test(mock, tc.testCase, TestConfig{ParallelOff: true})

			// Assert
			for _, lifecycle := range allLifecycles {
				timing, ok := td.timings[lifecycle]
				assert.True(t, ok)
				assert.Equal(t, lifecycle, timing.Lifecycle)
				assert.NotZero(t, timing.Start)
				assert.NotZero(t, timing.End)
				assert.NotZero(t, timing.Duration)
				if lifecycle == tc.wantLifecycle {
					// timed case
					assert.Equal(t, true, timing.Started)
					assert.Equal(t, true, timing.Ended)
					assert.True(t, int64(0) < timing.Duration.Nanoseconds())
				} else {
					// untimed case
					assert.Equal(t, false, timing.Started)
					assert.Equal(t, false, timing.Ended)
				}
			}
		})
	}
}

func Test_TestWithFatal_ShouldBeMarkedFatal(t *testing.T) {
	cases := map[string]struct {
		testCase      *TestCase
		wantLifecycle string
	}{
		constants.LifecycleArrange: {
			testCase: &TestCase{
				Arrange: func(t *TD) {
					t.Fatal()
				},
			},
			wantLifecycle: constants.LifecycleArrange,
		},
		constants.LifecycleAct: {
			testCase: &TestCase{
				Act: func(t *TD) {
					t.Fatal()
				},
			},
			wantLifecycle: constants.LifecycleAct,
		},
		constants.LifecycleAssert: {
			testCase: &TestCase{
				Assert: func(t *TD) {
					t.Fatal()
				},
			},
			wantLifecycle: constants.LifecycleAssert,
		},
		constants.LifecycleAfter: {
			testCase: &TestCase{
				After: func(t *TD) {
					t.Fatal()
				},
			},
			wantLifecycle: constants.LifecycleAfter,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			mock := newMockT()

			// Act
			td := Test(mock, tc.testCase, TestConfig{ParallelOff: true})

			// Assert
			assert.Equal(t, constants.Status{
				Status:    constants.StatusFail,
				Lifecycle: tc.wantLifecycle,
				Fatal:     true,
			}, td.statuses[0])
			assert.Equal(t, constants.Status{
				Status:    constants.StatusFail,
				Lifecycle: constants.LifecycleTestFinished,
				Fatal:     true,
			}, td.statuses[1])
		})
	}
}

func Test_TestWithErrorAndFatal_ShouldBeMarkedFatalAfterFatal(t *testing.T) {
	// Arrange
	mock := newMockT()

	// Act
	td := Test(mock, &TestCase{
		Act: func(t *TD) {
			t.Error()
			t.Fatal()
		},
	}, TestConfig{ParallelOff: true})

	// Assert
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleAct,
		Fatal:     false,
	}, td.statuses[0])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleAct,
		Fatal:     true,
	}, td.statuses[1])
	assert.Equal(t, constants.Status{
		Status:    constants.StatusFail,
		Lifecycle: constants.LifecycleTestFinished,
		Fatal:     true,
	}, td.statuses[2])
}

func Test_TestingT_RunShouldPass(t *testing.T) {
	// Arrange
	test := &TestCase{}
	test.Act = func(t *TD) {}
	name := Fname()

	// Act
	test.Run(t, name)

	// Assert
	assert.False(t, t.Failed())
}
