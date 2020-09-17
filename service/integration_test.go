package service_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/pkg/errors"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"

	"github.com/mercari/testdeck/service/config"
)

const goBinary = "go"

var defaultArgs = []string{
	"test",
	"-timeout",
	"30s",
	"-v",
	"-count=1", // never use cached results
}

func execute(pkg string) ([]byte, error) {
	args := append(defaultArgs, pkg)
	return exec.Command(goBinary, args...).CombinedOutput()
}

var kubernetesJobEnv = func(t *testing.T) {
	err := os.Setenv("RUN_AS", config.RunAsJob)
	if err != nil {
		t.Fatal("Problem setting up environment:", err)
	}
}

var localEnv = func(t *testing.T) {
	var err error
	err = os.Unsetenv("RUN_AS")
	if err != nil {
		t.Fatal("Problem setting up environment:", err)
	}
	err = os.Unsetenv(config.KubernetesServiceHostKey)
	if err != nil {
		t.Fatal("Problem setting up environment:", err)
	}
}

func TestIntegration(t *testing.T) {
	cases := map[string]struct {
		setupEnv     func(t *testing.T)
		pkg          string
		wantStrings  []string
		checkExecErr func(t *testing.T, err error)
	}{
		"LocalSuccess": {
			setupEnv: localEnv,
			pkg:      "github.com/mercari/testdeck/service/unit_tests/success",
			wantStrings: []string{
				"=== RUN   TestStub",
				"--- PASS: TestStub",
				"stub should pass",
			},
		},
		"Success": {
			setupEnv: kubernetesJobEnv,
			pkg:      "github.com/mercari/testdeck/service/unit_tests/success",
			wantStrings: []string{
				"PASS: TestStub",
				"stub should pass",
			},
		},
		"Statistics": {
			setupEnv: kubernetesJobEnv,
			pkg:      "github.com/mercari/testdeck/service/unit_tests/stats",
			wantStrings: []string{
				"PASS: TestStatistics",
				"stub should pass",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.setupEnv != nil {
				tc.setupEnv(t)
			}

			output, err := execute(tc.pkg)
			if tc.checkExecErr == nil && err != nil {
				t.Error(errors.Wrap(err, "Failed to execute test"))
			}
			if tc.checkExecErr != nil {
				tc.checkExecErr(t, err)
			}

			sout := string(output)
			t.Log(sout)

			for _, wantString := range tc.wantStrings {
				if !strings.Contains(sout, wantString) {
					t.Errorf("output doesn't contain: %s", wantString)
				}
			}
		})
	}
}
