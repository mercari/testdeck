package db

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mercari/testdeck/constants"
)

func Test_MySQLTimeZone(t *testing.T) {
	year := 2006
	day := 2
	month := time.January
	hour := 15
	min := 4
	sec := 5
	zone, _ := time.LoadLocation("Asia/Tokyo")
	td := time.Date(year, month, day, hour, min, sec, 0, zone)
	tdu := td.UTC()
	mst := MySQLTime{td}

	bs, err := mst.MarshalJSON()
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	re := regexp.MustCompile(`^"(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})"$`)
	n := 6
	matches := re.FindAllSubmatch(bs, n)
	if n != len(matches[0])-1 {
		t.Fatalf("Time format is incorrect. Marshalled string: %s", string(bs))
	}

	gotYear, _ := strconv.Atoi(string(matches[0][1]))
	gotMonth, _ := strconv.Atoi(string(matches[0][2]))
	gotDay, _ := strconv.Atoi(string(matches[0][3]))
	gotHour, _ := strconv.Atoi(string(matches[0][4]))
	gotMin, _ := strconv.Atoi(string(matches[0][5]))
	gotSec, _ := strconv.Atoi(string(matches[0][6]))
	if want, got := tdu.Year(), gotYear; want != got {
		t.Errorf("want year: %d, got: %d", want, got)
	}
	if want, got := int(tdu.Month()), gotMonth; want != got {
		t.Errorf("want month: %d, got: %d", want, got)
	}
	if want, got := tdu.Day(), gotDay; want != got {
		t.Errorf("want year: %d, got: %d", want, got)
	}
	if want, got := tdu.Hour(), gotHour; want != got {
		t.Errorf("want hour: %d, got: %d", want, got)
	}
	if want, got := tdu.Minute(), gotMin; want != got {
		t.Errorf("want minute: %d, got: %d", want, got)
	}
	if want, got := tdu.Second(), gotSec; want != got {
		t.Errorf("want second: %d, got: %d", want, got)
	}
}

type structWithID struct{ ID int }
type structWithoutID struct{}

func Test_SetIDWithStructPtrWithID_ShouldSucceed(t *testing.T) {
	p := &structWithID{}
	id := 1

	err := setID(p, id)

	if want := (error)(nil); want != err {
		t.Errorf("want: %v, got: %v", want, err)
	}

	if p.ID != id {
		t.Errorf("want: %d, got: %d", id, p.ID)
	}
}

func Test_SetIDWithStructPtrWithoutID_ShouldSucceed(t *testing.T) {
	p := &structWithoutID{}
	id := 1

	err := setID(p, id)

	if want := (error)(nil); want != err {
		t.Errorf("want: %v, got: %v", want, err)
	}
}

func Test_SetIDWithStructWithID_ShouldError(t *testing.T) {
	p := structWithoutID{}
	id := 1

	err := setID(p, id)

	if dontWant := (error)(nil); dontWant == err {
		t.Errorf("didn't want: %v, but got: %v", dontWant, err)
	}

	if want, got := "expects a pointer,", err.Error(); !strings.Contains(got, want) {
		t.Errorf("want substring: '%s', string contents: '%s'", want, got)
	}
}

func Test_SetIDWithNonStructPtr_ShouldError(t *testing.T) {
	p := "str"
	id := 1

	err := setID(&p, id)

	if dontWant := (error)(nil); dontWant == err {
		t.Errorf("didn't want: %v, but got: %v", dontWant, err)
	}

	if want, got := "expects a pointer to a struct,", err.Error(); !strings.Contains(got, want) {
		t.Errorf("want substring: '%s', string contents: '%s'", want, got)
	}
}

func Test_Save_ShouldSaveToActualDb(t *testing.T) {
	GuardProduction(t)
	os.Setenv(EnvKeyJobName, "test_job")
	os.Setenv(EnvKeyPodName, "test_job-pod")
	end := time.Now()
	mid := end.Add(-time.Second)
	start := end.Add(-time.Second * 2)
	s := []constants.Statistics{
		{
			Name:     t.Name(),
			Failed:   true,
			Fatal:    true,
			Start:    start,
			End:      end,
			Duration: end.Sub(start),
			Output:   "output goes here",
			Statuses: []constants.Status{
				{
					Status:    constants.StatusFail,
					Lifecycle: constants.LifecycleArrange,
					Fatal:     false,
				},
				{
					Status:    constants.StatusFail,
					Lifecycle: constants.LifecycleAct,
					Fatal:     true,
				},
				{
					Status:    constants.StatusFail,
					Lifecycle: constants.LifecycleTestFinished,
					Fatal:     true,
				},
			},
			Timings: map[string]constants.Timing{
				constants.LifecycleArrange: {
					Lifecycle: constants.LifecycleArrange,
					Start:     start,
					End:       mid,
					Duration:  mid.Sub(start),
					Started:   true,
					Ended:     true,
				},
				constants.LifecycleAct: {
					Lifecycle: constants.LifecycleAct,
					Start:     mid,
					End:       end,
					Duration:  end.Sub(mid),
					Started:   true,
					Ended:     true,
				},
			},
		},
	}
	g := New("go-test")
	jobID, jobErr := g.SaveJobStart()
	if jobErr != nil {
		t.Fatalf("Could not create initial job record, got: %v", jobErr)
	}

	IDs, err := g.Save(jobID, s)

	if wantErr := (error)(nil); wantErr != err {
		t.Errorf("Wanted error: %v, got error: %v", wantErr, err)
	}
	t.Logf("Saved results IDs: %v", IDs)
	if want, got := 1, len(IDs); want != got {
		t.Fatalf("Wanted len IDs: %d, got: %d", want, got)
	}
}
