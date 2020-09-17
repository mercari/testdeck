package unit_tests

import (
	"os"
	"testing"

	"github.com/mercari/testdeck"
	"github.com/mercari/testdeck/service"
)

func TestMain(m *testing.M) {
	os.Exit(service.Start(m))
}

func TestStub(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Log("pass!")
		},
	})
}
