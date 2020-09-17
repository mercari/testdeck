package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"os"
)

/*
config.go: Configurations for the GRPC microservice to stand up for testing
*/

const (
	envDevelopment = "development"
	RunAsLocal     = "local" // Running in local will not save test results to the DB
	RunAsJob       = "job"   // Set run_as ENV to "job" if you want to save test results to the DB
)

var (
	KubernetesServiceHostKey = "KUBERNETES_SERVICE_HOST"
	TelepresenceRootKey      = "TELEPRESENCE_ROOT"
)

// Env stores configuration settings extract from environmental variables
// by using https://github.com/kelseyhightower/envconfig
//
// The practice getting from environmental variables comes from https://12factor.net.
type Env struct {
	// Env is environment where application is running. This value is used to
	// annotate datadog metrics or sentry error reporting. The value should always
	// be "development"
	Env string `envconfig:"ENV" default:"development"`

	// GCPProjectID is you service GCP project ID
	GCPProjectID string `envconfig:"GCP_PROJECT_ID"`

	// Whether or not to print test realtime output to the event log
	PrintOutputToEventLog bool `envconfig:"PRINT_OUTPUT_TO_EVENT_LOG" default:"false"`

	// RunAs is an override to force testdeck to run as local tests or gRPC service
	// If RunAs is set as "service" then testdeck-service will always run as a gRPC service.
	// If RunAs is set as "local" then testdeck-service will always run as a normal local test.
	// For other values testdeck-service will fallback to its automatic sensing logic:
	//  - kubernetes with no teleprecense: gRPC service
	//  - otherwise: local
	RunAs string `envconfig:"RUN_AS"`

	// The URL of the DB to save test results to. If not declared, tests will still run but results can only be viewed through Kubernetes pod logs
	DbUrl string `envconfig:"DB_URL"`
}

func (e *Env) validate() error {
	checks := []struct {
		bad    bool
		errMsg string
	}{
		{
			e.Env != envDevelopment,
			fmt.Sprintf("invalid env is specified: %q", e.Env),
		},

		// Add your own validation here
	}

	for _, check := range checks {
		if check.bad {
			return errors.Errorf(check.errMsg)
		}
	}

	return nil
}

// ReadFromEnv reads configuration from environmental variables defined by Env struct
func ReadFromEnv() (*Env, error) {
	var env Env
	if err := envconfig.Process("", &env); err != nil {
		return nil, errors.Wrap(err, "failed to process envconfig")
	}

	if err := env.validate(); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	return &env, nil
}

// Detects if test service is running in a Kubernetes pod
func RunningInKubernetes() bool {
	_, set := os.LookupEnv(KubernetesServiceHostKey)
	return set
}

// Detects if test service is running in Telepresence
func RunningInTelepresence() bool {
	_, set := os.LookupEnv(TelepresenceRootKey)
	return set
}

// RunAs returns a string containing the type of environment the tests should run as.
//
// - service: run as a gRPC service pod (save results to database)
// - local: run using the standard Go runner (do not save results to database)
// - job: run as a k8 job (run-once, save results to database)
//
func RunAs(env *Env) string {
	if env.RunAs != "" {
		if env.RunAs == RunAsJob {
			return RunAsJob
		}

		// TODO emit warning if not actually set to "local"
		return RunAsLocal
	}

	// auto-detect environment tests are being run in
	if RunningInKubernetes() && !RunningInTelepresence() {
		// running in Kubernetes, not using Telepresence = regular test run job
		return RunAsJob
	}

	// default to local
	return RunAsLocal
}
