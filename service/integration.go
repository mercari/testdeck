package service

import (
	"fmt"
	"testing"

	"github.com/mercari/testdeck/runner"
	"github.com/mercari/testdeck/service/config"
	"github.com/mercari/testdeck/service/controller"
	"github.com/mercari/testdeck/service/db"
	"github.com/pkg/errors"
)

/*
service.go: Stands up a GRPC microservice to run tests on
*/

var Env *config.Env

// Test service only contains a Start() method to stand up the service, and a test runner
type Service interface {
	Start(opt ...ServiceOptions) int
	Runner() runner.Runner
}

// Controller allows the service to control the test run
type ServiceImpl struct {
	controller controller.Controller
}

// Creates a testdeck runner
func (s *ServiceImpl) Runner() runner.Runner {
	return s.controller.Runner()
}

// Additional configurations to pass as params
type ServiceOptions struct {
	RunOnceAndShutdown bool
}

func init() {
	var err error
	Env, err = config.ReadFromEnv()
	if err != nil {
		panic(errors.Wrap(err, "Could not read environment variables!"))
	}
}

// Starts up a GRPC microservice
func Start(m *testing.M, opt ...ServiceOptions) int {
	service := NewService(m)
	return service.Start(opt...)
}

// Creates a new service and DB client
func NewService(m *testing.M) Service {
	g := db.New(Env.GCPProjectID)
	return &ServiceImpl{
		controller: controller.New(m, g, Env),
	}
}

// Start the testing service
func (s *ServiceImpl) Start(opt ...ServiceOptions) int {
	s.controller.Runner().PrintOutputToEventLog(Env.PrintOutputToEventLog)
	
	// Only start tests if run as a test job
	switch runAs := config.RunAs(Env); runAs {
	case config.RunAsJob:
		s.controller.SetPrintToStdout(true)
		result := s.controller.RunAll()
		fmt.Println("\n" + result)
		return 0
	case config.RunAsLocal:
		fallthrough
	default:
		return s.controller.RunLocal()
	}
}
