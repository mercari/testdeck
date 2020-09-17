package testdeck

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// Test against the real testing.T by executing in sub commands.

const examplePackage = "github.com/mercari/testdeck/demo"
const goBinary = "go"

var defaultArgs = []string{"test", "-timeout", "30s", "-v", examplePackage, "-run"}

// Execute a "go test" on an example testcase and return the output.
// We have to do this because the testing.T package doesn't export a lot of
// methods required to use it directly.
func execute(testName string) ([]byte, error) {
	reTestName := fmt.Sprintf("^(%s)$", testName)
	args := append(defaultArgs, reTestName)
	return exec.Command(goBinary, args...).Output()
}

func Test_TestWithTestingT(t *testing.T) {
	cases := map[string]struct {
		testName        string
		wantStrings     []string // strings that should be in the expected output
		dontWantStrings []string // strings that should NOT be in the expected output
	}{
		"BasicExample_ActError": {
			testName: "Test_BasicExample_ActError",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_ActError",
				"basic act example error",
			},
		},
		"BasicExample_ActFatal": {
			testName: "Test_BasicExample_ActFatal",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_ActFatal",
				"basic act example fatal",
			},
		},
		"BasicExample_AssertErrorAndFatal": {
			testName: "Test_BasicExample_AssertErrorAndFatal",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_AssertErrorAndFatal",
				"basic assert example error",
				"basic assert example fatal",
			},
		},
		"BasicExample_ArrangeMultiFatal": {
			testName: "Test_BasicExample_ArrangeMultiFatal",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_ArrangeMultiFatal",
				"basic arrange example fatal 1",
			},
			dontWantStrings: []string{
				"basic arrange example fatal 2",
			},
		},
		"BasicExample_ArrangeAfterMultiFatal": {
			testName: "Test_BasicExample_ArrangeAfterMultiFatal",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_ArrangeAfterMultiFatal",
				"basic arrange example fatal 1",
				"basic after example fatal 2",
			},
		},
		"BasicExample_ArrangeFatalAndDeferred": {
			testName: "Test_BasicExample_ArrangeFatalAndDeferred",
			wantStrings: []string{
				"--- FAIL: Test_BasicExample_ArrangeFatalAndDeferred",
				"basic arrange example fatal 1",
				"basic deferred output",
			},
			dontWantStrings: []string{
				"basic arrange example fatal 2",
			},
		},
		"ActEmptyNoError": {
			testName: "Test_ActEmptyNoError",
			wantStrings: []string{
				"PASS: Test_ActEmptyNoError",
			},
		},
		"ArrangeSkipNow_ShouldbeMarkedSkip": {
			testName: "Test_ArrangeSkipNow_ShouldbeMarkedSkip",
			wantStrings: []string{
				"SKIP: Test_ArrangeSkipNow_ShouldbeMarkedSkip",
				"skip now defer message 1",
			},
			dontWantStrings: []string{
				"skip now defer message 2",
				"skip now act error",
				"skip now assert error",
				"skip now after message",
			},
		},
		"ActSkipNow_ShouldbeMarkedSkipAndExecuteAfter": {
			testName: "Test_ActSkipNow_ShouldbeMarkedSkipAndExecuteAfter",
			wantStrings: []string{
				"SKIP: Test_ActSkipNow_ShouldbeMarkedSkipAndExecuteAfter",
				"skip now after message",
				"skip now defer message",
			},
			dontWantStrings: []string{
				"skip now act error",
				"skip now assert error",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out, _ := execute(tc.testName)
			sout := string(out)

			for _, want := range tc.wantStrings {
				if !strings.Contains(sout, want) {
					t.Errorf("Want output substring: %s\n", want)
				}
			}

			for _, dontWant := range tc.dontWantStrings {
				if strings.Contains(sout, dontWant) {
					t.Errorf("Don't want output substring: %s\n", dontWant)
				}
			}
			t.Logf("\nCommand Test Output:\n%s\n", sout)
		})
	}
}
