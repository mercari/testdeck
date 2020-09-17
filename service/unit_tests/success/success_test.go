package success

import (
	"github.com/mercari/testdeck/service"
	"os"
	"testing"

	"github.com/mercari/testdeck"
	"github.com/mercari/testdeck/service/db"
)

func TestMain(m *testing.M) {
	os.Exit(service.Start(m, service.ServiceOptions{
		// make test code easier by shutting down immediately after one run
		RunOnceAndShutdown: true,
	}))
}

func TestStub(t *testing.T) {
	db.GuardProduction(t)
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			t.Log("stub should pass")
		},
	})
}
