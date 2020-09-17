# Introduction

Testdeck is a tool used for automating End-to-End (E2E) and Security tests for microservices.

## Concept & Architecture

Testdeck will test your deployed service from inside the development cluster, simulating the consumer-provider relationship shown below.

![Concept](images/concept.png?raw=true)

Architecture Summary:

- Each Service is paired with its own Test Service.
- The Test Service is deployed into a pod. It runs the test cases from within the cluster when it is deployed.
- (Optional) When the job finishes, the test results are saved to a database. A visual dashboard reads from the database and displays the test results. (See reporting_metrics.md for more information on how to set this up)

## Testdeck Lifecycle

Test stages execute in the following order:
0. FrameworkTestSetup (This step should never be used in your test cases. If tests fail at this step, it means that the framework failed prematurely for unexpected reasons)
1. Arrange
2. Act
3. Assert
4. After
5. FrameworkTestFinished (This step should never be used in your test cases. It is used to show that all steps have finished and results will be saved to the DB)
6. Deferred functions (declared in Arrange). These execute in the reverse order they were registered

After and Deferred functions are guaranteed to execute even on goexit (FailNow or SkipNow) because they are wrapped in a standard Golang defer function.

Note: You do not need to have all stages in your test case, you can omit stages that you don't need.

## Purpose

The target of this framework is integration testing, not unit testing (although it is built off of the unit testing framework go/testing). For
unit tests you should probably continue to use the standard Go testing
package.

For integration testing we need to achieve the following:

 - test with no mocks and ideally no fake services
 - properly detect flaky tests
 - classify the failure as best we can
 - easily perform the test consistently and repeatedly

For this reason we created a simple test harness.

## Dependencies

- [strethr/testify](https://github.com/stretchr/testify): To make the
framework's own self unit tests a bit shorter as well as prove that 3rd party
test package integration works
- [google/gofuzz](https://github.com/google/gofuzz): Integrated into the Fuzzer feature to generate random input values
- [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig): For retrieve environment variables for configuring the testing service
- [golang/go](https://github.com/golang/go): Testdeck is based off of Golang's native testing library

## Code Samples

Below are two samples of how to create a testdeck test case.

- In the **Inline Style**, you create a test case directly in the Test() function so that it will immediately run.
- In the **Struct Style**, you initialize the test case first, specify the stages, and then call Test() to start the test so that you can add additional actions before starting the test.

In most cases, either style is fine; it is just a matter of personal preferences.

```go
import "github.com/mercari/testdeck"

// Inline Style of writing tests
func TestInlineStyle_MathPow(t *testing.T) {
	var x, want, got float64

	// Create the test case directly inside Test()
	testdeck.Test(t, &testdeck.TestCase{
		Arrange: func(t *testdeck.TD) {
			x = 3.0
			want = 9.0
		},
		Act: func(t *testdeck.TD) {
			got = math.Pow(x, 2.0)
		},
		Assert: func(t *testdeck.TD) {
			if want != got {
				t.Errorf("want: %f, got %f", want, got)
			}

		},
	})
}

// Struct style of writing tests
func TestStructStyle_MathPow(t *testing.T) {
	var x, want, got float64

	// Initialize the test case first and then specify the stages
	test := testdeck.TestCase{}
	test.Arrange = func(t *testdeck.TD) {
)		x = 3.0
		want = 9.0
	}
	test.Act = func(t *testdeck.TD) {
		got = math.Pow(x, 2.0)
	}
	test.Assert = func(t *testdeck.TD) {
		if want != got {
			t.Errorf("want: %d, got: %d", want, got)
		}
	}

	// Finally, call Test() to start the test
	testdeck.Test(t, &test)
}
```

You can also reuse any setup code for your tests with the following strategy.

```go
func Setup(shared *int) func(t *testdeck.TD) {
	return func(t *testdeck.TD) {
		*shared = 42
	}
}

func TestReusableArrangeInline(t *testing.T) {
	var value int

	testdeck.Test(t, &testdeck.TestCase{
		// Setup() can be inserted into the test case here
		Arrange: Setup(&value),
		Assert: func(t *testdeck.TD) {
			if value != 42 {
				t.Errorf("this is not the meaning of life: %d", value)
			}
			t.Logf("The meaning of life! %d", value)
		},
	})
}
```

## How NOT to write tests

All steps (Arrange, Act, Assert, After) are optional, so the following test where everything is stuffed into one stage is technically possible but **discouraged**:

```go
// ❌ Do NOT do this! ❌
func Test_DiscouragedExample(t *testing.T) {
	testdeck.Test(t, &testdeck.TestCase{
		// why are you putting everything into the Act stage?
		Act: func(t *testdeck.TD) {
			want := &ServiceResponse{}
			token, err := PotentiallyFlakyTokenRetrieval()
			if err != nil {
				t.Fatal("failed to retieve token", err)
			}

			got := client.ServiceRequest(token)

			if want != got {
				t.Errorf("want: %f, got %f", want, got)
			}
			err := PotentiallyFlakyDatabaseCheck(got.ID)
			if err != nil {
				t.Fatal("failed to find ID in database", err)
			}
		},
	})
}
```

We do not recommend this because:

- we will not be able to detect flakiness in the proper test steps (e.g. Arrange)
- we cannot report useful information like execution time for the operation you want to test

For example, in the above test, if `PotentiallyFlakyTokenRetrieval` or
`PotentiallyFlakyDatabaseCheck` fails, then all we can report is that your
test failed and no further information.

Instead, we recommend rewriting the test like below:

```go
func Test_EncouragedExample(t *testing.T) {
	var want *ServiceResponse
	var token string

	testdeck.Test(t, &testdeck.TestCase{
		// Test code is neatly separated into stages
		Arrange: func(t *testdeck.TD) {
			want := &ServiceResponse{}
			token, err := PotentiallyFlakyTokenRetrieval()
			if err != nil {
				t.Fatal("failed to retieve token", err)
			}
		},
		Act: func(t *testdeck.TD) {
			got := client.ServiceRequest(token)
		},
		Assert: func(t *testdeck.TD) {
			if want != got {
				t.Errorf("want: %f, got %f", want, got)
			}
			err := PotentiallyFlakyDatabaseCheck(got.ID)
			if err != nil {
				t.Fatal("failed to find ID in database", err)
			}
		},
	})
}
```