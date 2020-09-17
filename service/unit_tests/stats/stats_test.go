package success

import (
	"github.com/mercari/testdeck/service"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mercari/testdeck"
	"github.com/mercari/testdeck/runner"
	"github.com/mercari/testdeck/service/db"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(service.Start(m, service.ServiceOptions{
		// make test code easier by shutting down immediately after one run
		RunOnceAndShutdown: true,
	}))
}

func TestStatistics(t *testing.T) {
	db.GuardProduction(t)
	wantPause := time.Millisecond * 10
	testdeck.Test(t, &testdeck.TestCase{
		Act: func(t *testdeck.TD) {
			time.Sleep(wantPause)
			t.Log("stub should pass")
		},
	})

	r := runner.Instance(nil)
	stats := r.Statistics()
	require.Equal(t, 1, len(stats))
	assert.NotNil(t, stats[0].Start)
	assert.NotNil(t, stats[0].End)
	assert.True(t, stats[0].Duration.Nanoseconds() > wantPause.Nanoseconds())
}
