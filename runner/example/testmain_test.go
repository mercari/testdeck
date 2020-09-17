package example

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mercari/testdeck/runner"
	"testing"

	"github.com/mercari/testdeck"
)

const pass = false // set to false to force the test to fail
var printTestOutput = flag.Bool("printTestOutput", false, "print normal Go test output to STDOUT")

func TestMain(m *testing.M) {
	flag.Parse()
	r := runner.Instance(m)
	r.PrintToStdout(*printTestOutput)
	r.Match("Stub")
	r.Run()
	stats, err := json.Marshal(r.Statistics())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(stats))
}

func Test_RunnerStub(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			if pass {
				t.Log("I will pass")
			} else {
				t.Log("I will fail")
				t.Error("fail requested")
			}
		},
	})
}

func Test_AnotherPassingStub(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Log("passing stub")
		},
	})
}

func Test_AnotherFailingStub(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Error("failing stub")
		},
	})
}
