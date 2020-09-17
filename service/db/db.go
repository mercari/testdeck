package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mercari/testdeck/service/config"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/mercari/testdeck/constants"
)

/*
db.go: This is an example of how you can save test results to a DB. Your organization may require different behavior and data so please feel free to fork/clone this repo and modify for your needs
*/

// -----
// Environment and Constants
// -----

// HTTPClient is the http.Client to use for all REST API requests
var HTTPClient *http.Client

const (
	EnvKeyJobName = "JOB_NAME"
	EnvKeyPodName = "POD_NAME"
)

const (
	// Job is the `job` Endpoint map key
	Job = "job"
	// JobUpdate is the `job` PUT Endpoint map key
	JobUpdate = "jobUpdate"
	// Statistic is the `result` Endpoint map key
	Statistic = "statistic"
	// Timing is the `timing` Endpoint map key
	Timing = "timing"
	// Status is the `status` Endpoint map key
	Status = "status"
)

// Examples of endpoints for accessing the test results DB
var Endpoints = map[string]string{
	Job:       "/job",
	JobUpdate: "/job/:id",
	Statistic: "/result",
	Timing:    "/timing",
	Status:    "/status",
}

type ServerResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

// Represents a DB client
type Db struct {
	GcpProjectID string
}

// -----
// Helper methods
// -----

func readFromConfig() *config.Env {
	env, err := config.ReadFromEnv()
	if err != nil {
		panic(errors.Wrap(err, "Could not read environment variables!"))
	}

	return env
}

// GuardProduction will cause a test to fail if the URL looks like a production
// URL. Use this as the first line in a Test to prevent accidentally saving
// library tests results in production DB.
func GuardProduction(t *testing.T) {
	url := readFromConfig().DbUrl
	if strings.Contains(strings.ToLower(url), "prod") {
		t.Fatalf("You are trying to run this test against what looks like a production database! Current DB URL: %s", url)
	}
}

func composeEndpoint(endpoint string) string {
	return readFromConfig().DbUrl + Endpoints[endpoint]
}

// MySQLTime is a wrapper on time.Time that will marshall to MySQL timestamp format for Json
type MySQLTime struct {
	time.Time
}

func (t MySQLTime) MarshalJSON() ([]byte, error) {
	ts := t.UTC().Format("\"2006-01-02 15:04:05\"")
	return []byte(ts), nil
}

// -----
// Saving jobs
// -----

type job struct {
	ID           int           `json:"id"`
	GcpProjectID string        `json:"gcp_project_id"`
	JobName      string        `json:"job_name"`
	PodName      string        `json:"pod_name"`
	Finished     bool          `json:"finished"`
	Failed       bool          `json:"failed"`
	Start        MySQLTime     `json:"start_ts"`
	End          MySQLTime     `json:"end_ts"`
	Duration     time.Duration `json:"duration_ns"`
}

func newJob() *job {
	jb := &job{}
	jb.JobName = os.Getenv(EnvKeyJobName)
	jb.PodName = os.Getenv(EnvKeyPodName)
	jb.Start = MySQLTime{time.Now()}
	jb.End = jb.Start
	return jb
}

type jobUpdate struct {
	Finished bool          `json:"finished"`
	Failed   bool          `json:"failed"`
	Start    MySQLTime     `json:"start_ts"`
	End      MySQLTime     `json:"end_ts"`
	Duration time.Duration `json:"duration_ns"`
}

// Writes the initial job record to DB and returns the row ID
// This marks the start of a test run
func (g *Db) SaveJobStart() (int, error) {
	jb := newJob()
	jb.GcpProjectID = g.GcpProjectID
	err := insertRestOperation(composeEndpoint(Job), jb)
	return jb.ID, err
}

// Update the created job record with the final results
// This marks the end of a test run
func (g *Db) updateJobRecord(ID int, stats []constants.Statistics) error {
	update := jobUpdate{
		Failed:   false,
		Finished: true,
	}
	for _, stat := range stats {
		if stat.Failed {
			update.Failed = true
		}
		// find actual starting time
		if update.Start.IsZero() || update.Start.After(stat.Start) {
			update.Start = MySQLTime{stat.Start}
		}
		// find actual ending time
		if stat.End.After(update.End.Time) {
			update.End = MySQLTime{stat.End}
		}
	}
	update.Duration = update.End.Sub(update.Start.Time)

	resource := strings.Replace(composeEndpoint(JobUpdate), ":id", strconv.Itoa(ID), -1)
	return updateRestOperation(resource, update)
}

// -----
// Saving test results
// -----

type result struct {
	ID           int           `json:"id"`
	JobID        int           `json:"job_id"`
	GcpProjectID string        `json:"gcp_project_id"`
	Name         string        `json:"test_name"`
	Failed       bool          `json:"failed"`
	Fatal        bool          `json:"fatal"`
	Start        MySQLTime     `json:"start_ts"`
	End          MySQLTime     `json:"end_ts"`
	Duration     time.Duration `json:"duration_ns"`
	Output       string        `json:"output_text"`
}

func newResultFrom(jobID int, stats constants.Statistics) *result {
	return &result{
		JobID:    jobID,
		Name:     stats.Name,
		Failed:   stats.Failed,
		Fatal:    stats.Fatal,
		Start:    MySQLTime{stats.Start},
		End:      MySQLTime{stats.End},
		Duration: stats.Duration,
		Output:   stats.Output,
	}
}

// Writes a set of data.Statistics to the DB database and returns the row numbers
func (g *Db) Save(jobID int, stats []constants.Statistics) (IDs []int, err error) {
	err = g.updateJobRecord(jobID, stats)
	if err != nil {
		return nil, err
	}

	IDs = []int{}
	for _, stat := range stats {
		resultID, err := g.saveStatistics(jobID, stat)
		if err != nil {
			return nil, err
		}
		IDs = append(IDs, resultID)
	}
	return IDs, nil
}

func (g *Db) saveStatistics(jobID int, stat constants.Statistics) (resultID int, err error) {
	resultID, err = g.saveStatisticsRow(jobID, stat)
	if err != nil {
		return 0, err
	}
	for _, t := range stat.Timings {
		err = g.saveTimingRow(t, resultID)
		if err != nil {
			return 0, err
		}
	}
	for _, s := range stat.Statuses {
		err = g.saveStatusRow(s, resultID)
		if err != nil {
			return 0, err
		}
	}
	return resultID, err
}

func (g *Db) saveStatisticsRow(jobID int, s constants.Statistics) (ID int, err error) {
	r := newResultFrom(jobID, s)
	r.GcpProjectID = g.GcpProjectID
	err = insertRestOperation(composeEndpoint(Statistic), r)
	ID = r.ID
	return
}

// -----
// Saving timing statistics
// -----

type timing struct {
	ResultsID int           `json:"results_id"`
	Lifecycle string        `json:"lifecycle_value"`
	Start     MySQLTime     `json:"start_ts"`
	End       MySQLTime     `json:"end_ts"`
	Duration  time.Duration `json:"duration_ns"`
	Started   bool          `json:"started_tc"`
	Ended     bool          `json:"ended_tc"`
}

func newTimingFrom(t constants.Timing, resultID int) *timing {
	return &timing{
		ResultsID: resultID,
		Lifecycle: t.Lifecycle,
		Start:     MySQLTime{t.Start},
		End:       MySQLTime{t.End},
		Duration:  t.Duration,
		Started:   t.Started,
		Ended:     t.Ended,
	}
}

func (g *Db) saveTimingRow(t constants.Timing, resultID int) error {
	return insertRestOperation(composeEndpoint(Timing), newTimingFrom(t, resultID))
}

// -----
// Saving test status
// -----

type status struct {
	ID        int    `json:"id"`
	ResultsID int    `json:"results_id"`
	Status    string `json:"status_value"`
	Lifecycle string `json:"lifecycle_value"`
	Fatal     bool   `json:"fatal"`
}

func newStatusFrom(s constants.Status, resultID int) *status {
	return &status{
		ResultsID: resultID,
		Status:    s.Status,
		Lifecycle: s.Lifecycle,
		Fatal:     s.Fatal,
	}
}

func (g *Db) saveStatusRow(s constants.Status, resultID int) error {
	return insertRestOperation(composeEndpoint(Status), newStatusFrom(s, resultID))
}

// -----
// Methods
// -----

// Creates a new DB client to access the DB
func New(gcpProjectID string) *Db {
	return &Db{
		GcpProjectID: gcpProjectID,
	}
}

func initHTTPClient() {
	HTTPClient = &http.Client{
		Timeout: constants.DefaultHttpTimeout,
	}
}

// Creates a POST request
func insertRestOperation(resource string, payload interface{}) error {
	j, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error during insert JSON marshalling")
	}

	request, err := http.NewRequest("POST", resource, bytes.NewBuffer(j))
	if err != nil {
		return errors.Wrap(err, "error building POST request")
	}
	request.Header.Set("Content-Type", "application/json")

	if HTTPClient == nil {
		initHTTPClient()
	}

	response, err := HTTPClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "error doing insert HTTP request")
	}

	responseBody := &bytes.Buffer{}
	_, err = responseBody.ReadFrom(response.Body)
	if err != nil {
		return errors.Wrap(err, "error reading insert response body")
	}

	sr := ServerResponse{}
	err = json.Unmarshal(responseBody.Bytes(), &sr)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling insert response")
	}

	if sr.Status == "error" {
		return fmt.Errorf("DB Server Error: %s", sr.Error)
	}

	if sr.ID != 0 {
		// save the ID to the payload if possible
		setID(payload, sr.ID)
	}

	return nil
}

// Creates a PUT request
func updateRestOperation(resource string, payload interface{}) error {
	j, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error during update JSON marshalling")
	}

	request, err := http.NewRequest("PUT", resource, bytes.NewBuffer(j))
	if err != nil {
		return errors.Wrap(err, "error building PUT request")
	}
	request.Header.Set("Content-Type", "application/json")

	if HTTPClient == nil {
		initHTTPClient()
	}

	response, err := HTTPClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "error doing update HTTP request")
	}

	responseBody := &bytes.Buffer{}
	_, err = responseBody.ReadFrom(response.Body)
	if err != nil {
		return errors.Wrap(err, "error reading update response body")
	}

	sr := ServerResponse{}
	err = json.Unmarshal(responseBody.Bytes(), &sr)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling update response")
	}

	if sr.Status == "error" {
		return fmt.Errorf("DB"+
			" Server Error: %s", sr.Error)
	}

	return nil
}

/*
 Set the ID of a pointer to a struct if it contains an ID int field. If the
 struct pointer contains no ID field, do nothing.

 payload should be a pointer to a struct, otherwise an error will be
 returned.
*/
func setID(payload interface{}, id int) error {
	v := reflect.ValueOf(payload)
	if k := v.Kind(); k != reflect.Ptr {
		return fmt.Errorf("setID expects a pointer, instead got Kind: %v", k)
	}

	e := v.Elem()
	if k := e.Type().Kind(); k != reflect.Struct {
		return fmt.Errorf("setID expects a pointer to a struct, instead got Kind: %v", k)
	}

	field := e.FieldByName("ID")
	if field.IsValid() && field.CanSet() && field.Kind() == reflect.Int {
		field.SetInt(int64(id))
	}

	return nil
}
