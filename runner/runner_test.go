package runner

import (
	"regexp"
	"testing"

	"github.com/mercari/testdeck/constants"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Runner_ShouldAddStatistics(t *testing.T) {
	// Arrange
	bm := badM{}
	r := newInstance(&bm)
	stats := constants.Statistics{Name: "foobar"}

	// Act
	r.AddStatistics(&stats)

	// Assert
	got := r.Statistics()
	require.Equal(t, 1, len(got))
	assert.Equal(t, stats, got[0])
}

// This test consumes the "singleton" behavior of the file. Because of this
// behavior, we have to move some assertions ahead to guarantee we valid state
// before initialization.
func Test_Runner_ShouldInitialize(t *testing.T) {
	// Arrange
	deps := &TestDeps{}
	m := testing.MainStart(deps, make([]testing.InternalTest, 0), make([]testing.InternalBenchmark, 0), make([]testing.InternalExample, 0))
	assert.False(t, Initialized()) // check before initialization

	// Act
	_ = Instance(m)

	// Assert
	assert.True(t, Initialized())
}

func Test_Runner_ShouldFindInternalTestField(t *testing.T) {
	// Arrange
	itest := testing.InternalTest{
		Name: "me",
		F: func(t *testing.T) {
			t.Log("hi")
		},
	}
	itests := []testing.InternalTest{itest}
	deps := &TestDeps{}
	m := testing.MainStart(deps, itests, make([]testing.InternalBenchmark, 0), make([]testing.InternalExample, 0))

	// Act
	got := getInternalTests(m)

	// Assert
	assert.Equal(t, len(itests), len(got))
	assert.True(t, len(got) > 0)
	assert.Equal(t, itest.Name, got[0].Name)
}

type badM struct {
	dummyStr      string
	dummyIntSlice []int
}

func (b *badM) Run() int {
	return 0
}

func Test_Runner_ShouldNotFindInternalTestAndPanic(t *testing.T) {
	// Arrange
	bm := badM{}
	defer func() {
		r := recover()
		if r == nil {
			// Assert
			t.Error("getInternalTests did not panic when it should")
		}
	}()

	// Act
	getInternalTests(&bm)

	// Assert
}

func Test_FilterTest_ShouldFilterTests(t *testing.T) {
	names := []string{"A", "AA", "AAA"}
	testFunc := func(t *testing.T) {}
	var internalTests []testing.InternalTest
	for _, name := range names {
		internalTests = append(internalTests, testing.InternalTest{
			F:    testFunc,
			Name: name,
		})
	}

	tests := map[string]struct {
		re        *regexp.Regexp
		wantNames []string
	}{
		"All": {
			re:        regexp.MustCompile(".*"),
			wantNames: names,
		},
		"AAA": {
			re:        regexp.MustCompile("AAA"),
			wantNames: []string{"AAA"},
		},
		"AA": {
			re:        regexp.MustCompile("AA"),
			wantNames: []string{"AA", "AAA"},
		},
		"^A$": {
			re:        regexp.MustCompile("^A$"),
			wantNames: []string{"A"},
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			// Act
			filtered := filterTests(tc.re, internalTests)

			// Assert
			assert.Equal(t, len(tc.wantNames), len(filtered))
			var filteredNames []string
			for _, ftest := range filtered {
				filteredNames = append(filteredNames, ftest.Name)
			}
			for _, wantName := range tc.wantNames {
				assert.Contains(t, filteredNames, wantName)
			}
		})
	}
}

func Test_FilterTestWorkaround_ShouldTagTestNames(t *testing.T) {
	// Arrange
	pattern := "^AAA$"
	re := regexp.MustCompile(pattern)
	names := []string{
		"A",
		"AA",
		"AAA",
	}
	wantNames := []string{
		pattern + "\x00A",
		pattern + "\x00AA",
		pattern + "\x00AAA",
	}
	testFunc := func(t *testing.T) {}
	var internalTests []testing.InternalTest
	for _, name := range names {
		internalTests = append(internalTests, testing.InternalTest{
			F:    testFunc,
			Name: name,
		})
	}

	// Act
	filtered := filterTestsWorkaround(re, internalTests, true, pattern)

	// Assert
	assert.Equal(t, len(names), len(filtered))
	var filteredNames []string
	for _, ftest := range filtered {
		filteredNames = append(filteredNames, ftest.Name)
	}
	assert.ElementsMatch(t, wantNames, filteredNames)
}

func Test_MatchTag_ShouldMatchTaggedName(t *testing.T) {
	// Arrange
	name := "^AAA$\x00AAA"

	// Act
	tagged, matched, actual := MatchTag(name)

	// Assert
	assert.True(t, tagged)
	assert.True(t, matched)
	assert.Equal(t, actual, "AAA")
}

func Test_MatchTag_ShouldNotMatchWithTaggedNameNotMatchingPattern(t *testing.T) {
	// Arrange
	name := "^AA$\x00AAA"

	// Act
	tagged, matched, actual := MatchTag(name)

	// Assert
	assert.True(t, tagged)
	assert.False(t, matched)
	assert.Equal(t, actual, "AAA")
}

func Test_MatchTag_ShouldNotMatchWithUntaggedName(t *testing.T) {
	// Arrange
	name := "AAA/aaa"

	// Act
	tagged, matched, actual := MatchTag(name)

	// Assert
	assert.False(t, tagged)
	assert.False(t, matched)
	assert.Equal(t, name, actual)
}

type FakeM struct {
	t         *testing.T
	deps      *TestDeps
	mockTests []testing.InternalTest
}

// Run emulates what the testing library would do for us.
func (m *FakeM) Run() int {
	for _, test := range m.mockTests {
		match, err := m.deps.MatchString("", test.Name)
		if err != nil {
			panic(errors.Wrap(err, "MatchString failed."))
		}
		if match {
			test.F(m.t)
		}
	}
	return 0
}

// Test theorhetically should work but the testing package will panic because it
// can't close an already close testlog file. Tests are left for reference but
// should not be run due to this issue. Perhaps this is a good indication of how
// fragile this runner is.
func test_Runner(t *testing.T) {
	tests := []testing.InternalTest{
		testing.InternalTest{
			Name: "TestPass1",
			F: func(t *testing.T) {
				t.Log("pass1")
			},
		},
		testing.InternalTest{
			Name: "TestPass2",
			F: func(t *testing.T) {
				t.Log("pass2")
			},
		},
		testing.InternalTest{
			Name: "TestFail",
			F: func(t *testing.T) {
				t.Error("fail")
			},
		},
		testing.InternalTest{
			Name: "TestFatal",
			F: func(t *testing.T) {
				t.Fatal("fatal")
			},
		},
	}

	cases := map[string]struct {
		pattern     string
		wantStrings []string
	}{
		"SinglePass": {
			pattern: "TestPass1",
			wantStrings: []string{
				"PASS: TestPass1",
				"pass1",
			},
		},
		"MultiplePass": {
			pattern: "TestPass",
			wantStrings: []string{
				"PASS: TestPass1",
				"pass1",
				"PASS: TestPass2",
				"pass2",
			},
		},
		"Fail": {
			pattern: "TestFail",
			wantStrings: []string{
				"FAIL: TestFail",
				"fail",
			},
		},
		"Fatal": {
			pattern: "TestFatal",
			wantStrings: []string{
				"FAIL: TestFatal",
				"fatal",
			},
		},
		"MultipleFail": {
			pattern: "TestFa",
			wantStrings: []string{
				"FAIL: TestFail",
				"fail",
				"FAIL: TestFatal",
				"fatal",
			},
		},
	}

	deps := &TestDeps{}
	fm := &FakeM{
		t:         t,
		deps:      deps,
		mockTests: tests,
	}
	r := newInstance(fm)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			r.Match(tc.pattern)

			// Act
			r.Run()

			// Assert
			for _, wantString := range tc.wantStrings {
				assert.Contains(t, r.Output(), wantString)
			}
		})
	}
}
