package config

import (
	"os"
	"testing"
)

const (
	testGCPProjectID = "testdeck-service"
)

func TestReadFromEnv(t *testing.T) {
	reset := setenvs(t, map[string]string{
		"ENV":                       envDevelopment,
		"GCP_PROJECT_ID":            testGCPProjectID,
		"PRINT_OUTPUT_TO_EVENT_LOG": "true",
	})
	defer reset()

	env, err := ReadFromEnv()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if got, want := env.GCPProjectID, testGCPProjectID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	if got, want := env.PrintOutputToEventLog, true; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReadFromEnvValidationFailed(t *testing.T) {
	reset := setenvs(t, map[string]string{
		"ENV":            "prod",
		"GCP_PROJECT_ID": testGCPProjectID,
	})
	defer reset()

	_, err := ReadFromEnv()
	if err == nil {
		t.Fatalf("expect to be failed")
	}
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		env     *Env
		success bool
	}{
		"Valid1": {
			&Env{
				Env: envDevelopment,
			},
			true,
		},

		"InvalidEnv": {
			&Env{
				Env: "staging",
			},
			false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.env.validate()
			if err != nil {
				if tc.success {
					t.Fatalf("expect not to be failed: %s", err)
				}
				return
			}

			if !tc.success {
				t.Fatalf("expect to be failed")
			}
		})
	}
}

func TestRunningInKubernetes(t *testing.T) {
	setenv(t, KubernetesServiceHostKey, "abc")

	if !RunningInKubernetes() {
		t.Error("Should be running in kubernetes!")
	}
}

func TestNotRunningInKubernetes(t *testing.T) {
	unsetenv(t, KubernetesServiceHostKey)

	if RunningInKubernetes() {
		t.Error("Should not be running in kubernetes!")
	}
}

func TestRunAs(t *testing.T) {
	setEnvAsLocal := func() {
		unsetenv(t, KubernetesServiceHostKey)
		unsetenv(t, TelepresenceRootKey)
	}

	setEnvAsLocalTelepresence := func() {
		setenv(t, KubernetesServiceHostKey, "host")
		setenv(t, TelepresenceRootKey, "root")
	}

	setEnvAsKubernetes := func() {
		setenv(t, KubernetesServiceHostKey, "host")
		unsetenv(t, TelepresenceRootKey)
	}

	cases := map[string]struct {
		setup func()
		env   *Env
		want  string
	}{
		"Local": {
			setup: setEnvAsLocal,
			env:   &Env{RunAs: ""},
			want:  RunAsLocal,
		},
		"LocalGarbage": {
			setup: setEnvAsLocal,
			env:   &Env{RunAs: "garbage-value"},
			want:  RunAsLocal,
		},
		"LocalForcedKubernetes": {
			setup: setEnvAsKubernetes,
			env:   &Env{RunAs: "local"},
			want:  RunAsLocal,
		},
		"LocalForcedKubernetesGarbage": {
			setup: setEnvAsKubernetes,
			env:   &Env{RunAs: "garbage-value"},
			want:  RunAsLocal,
		},
		"LocalTelepresence": {
			setup: setEnvAsLocalTelepresence,
			env:   &Env{RunAs: ""},
			want:  RunAsLocal,
		},
		"Job": {
			setup: setEnvAsKubernetes,
			env:   &Env{RunAs: ""},
			want:  RunAsJob,
		},
		"JobForced": {
			setup: setEnvAsLocal,
			env:   &Env{RunAs: "job"},
			want:  RunAsJob,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			if tc.setup != nil {
				tc.setup()
			}

			// Act
			got := RunAs(tc.env)

			// Assert
			if tc.want != got {
				t.Errorf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func setenv(t *testing.T, k, v string) func() {
	t.Helper()

	prev := os.Getenv(k)
	if err := os.Setenv(k, v); err != nil {
		t.Fatal(err)
	}

	return func() {
		if prev == "" {
			os.Unsetenv(k)
		} else {
			if err := os.Setenv(k, prev); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func unsetenv(t *testing.T, k string) func() {
	t.Helper()

	prev := os.Getenv(k)
	if err := os.Unsetenv(k); err != nil {
		t.Fatal(err)
	}

	return func() {
		if prev == "" {
			return
		} else {
			if err := os.Setenv(k, prev); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func setenvs(t *testing.T, kv map[string]string) func() {
	t.Helper()

	resetFs := make([]func(), 0, len(kv))
	for k, v := range kv {
		resetF := setenv(t, k, v)
		resetFs = append(resetFs, resetF)
	}

	return func() {
		for _, resetF := range resetFs {
			resetF()
		}
	}
}
