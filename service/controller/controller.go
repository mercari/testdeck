package controller

import (
	"fmt"
	"log"
	"testing"

	"github.com/mercari/testdeck/constants"

	"github.com/mercari/testdeck/runner"
	"github.com/mercari/testdeck/service/config"
	"github.com/mercari/testdeck/service/db"
)

/*
controller.go: This contains methods for controlling the test run (Run methods, test runner, etc.)
*/

// Interface for the controller of the test run
type Controller interface {
	RunLocal() int
	RunAll() string
	Run(name string) (constants.Statistics, []int, bool) // Note: This is currently not being used because running individual tests by test name is not supported yet
	Runner() runner.Runner
	SetPrintToStdout(bool)
}

// Implements the interface above
type controllerImpl struct {
	m      *testing.M
	runner runner.Runner
	g      *db.Db
	env    *config.Env
}

// Creates a new test controller
func New(m *testing.M, g *db.Db, env *config.Env) Controller {
	runner := runner.Instance(m)
	// disable stdout for test output since we are running as a gRPC microservice
	runner.PrintToStdout(false)
	controller := &controllerImpl{
		m:      m,
		runner: runner,
		g:      g,
		env:    env,
	}
	runner.SetEventLogger(controller)
	return controller
}

// Creates a test runner
func (c *controllerImpl) Runner() runner.Runner {
	return c.runner
}

// Set test run mode as local
func (c *controllerImpl) RunLocal() int {
	return c.m.Run()
}

// Logs an event
func (c *controllerImpl) Log(message string) {
	// FIXME stdout is captured by the test log
	log.Printf("Runner Event: %s", message)
}

// Set output to stdout
func (c *controllerImpl) SetPrintToStdout(printToStdout bool) {
	c.runner.PrintToStdout(printToStdout)
}

// -----
// Run Methods
// -----

// Runs a set of tests matching the regex pattern
func (c *controllerImpl) runSet(pattern string) (IDs []int, savedToDb bool, stats []constants.Statistics) {
	var err error
	var jobID int
	env, _ := config.ReadFromEnv()

	// If running as a job and a DB URL was declared, enable save to DB functionality
	saveToDatabase := config.RunAs(c.env) == config.RunAsJob && env.DbUrl != ""
	if saveToDatabase {
		jobID, err = c.g.SaveJobStart()
		if err != nil {
			// TODO
			log.Printf("Warning: failed to save results to database: %v", err)
		} else {
			log.Printf("Job ID: %d", jobID)
		}
	}

	// Run all tests matching the pattern
	c.runner.Match(pattern)
	c.runner.Run()

	// Save statistics to DB
	stats = c.runner.Statistics()
	c.runner.ClearStatistics()
	if len(stats) > 0 && saveToDatabase {
		var err error
		IDs, err = c.g.Save(jobID, stats)
		if err != nil {
			log.Printf(
				"Warning: failed to save results to database: %v", err)
		} else {
			log.Printf("Test IDs: %v", IDs)
			savedToDb = true
		}
	}

	return
}

// Run all tests
func (c *controllerImpl) RunAll() string {
	_, _, stats := c.runSet(".*")

	for _, stat := range stats {
		if stat.Failed {
			return constants.ResultFail
		}
	}
	return constants.ResultPass
}

// Run an individual test case by name
// Note: This method isn't used anywhere right now because the feature to run individual tests by name is not complete yet
func (c *controllerImpl) Run(name string) (constants.Statistics, []int, bool) {
	IDs, savedToDb, stats := c.runSet(fmt.Sprintf("^%s$", name))

	if len(stats) == 0 {
		// TODO: Add logic for when no matching test cases are found
		return constants.Statistics{}, []int{}, false
	}

	return stats[0], IDs, savedToDb
}
