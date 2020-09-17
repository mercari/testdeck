# Intruder

The Intruder, similar to Burpsuite's Intruder function, feeds malicious payloads into all parameters of a grpc endpoint.

Relevant files:
- intruder
    - intruder.go
    - testdata_helper.go

## How to Use

Pass the following parameters into `RunIntruderTests()`:
- t (the current test case)
- context
- the testdeck.TestCase
- grpc client of the service
- the rpc method to test
- a valid, sample request (this will act as the base for the intruder)
- the json data set (explained below)

```
var sampleRequest = &pb.SayRequest{MessageId: "test", MessageBody: "test"}

func Test_Say_SQLiIntruderTest(t *testing.T) {
	var client interface{}

	var testDataSet InputValidationTestData

	tc := testdeck.TestCase{}

	// Arrange
	test.Arrange = func(t *testdeck.TD) {
		// set up your client here

		// read in the fuzz test data from a json file
		testDataSet, _ = ParseInputValidationTestDataFromJson("../payloads/sql_injection/testdata.json")

		// do other set up steps here
	}

	// fuzz using the sample request, fuzzing function, and specified json data file
	RunIntruderTests(t, context.TODO(), tc, client, "Say", sampleRequest, testDataSet)
}
```

The output will look similar to below:

```
=== RUN   Test_Say_SQLiIntruderTest/MessageBody
    Test_Say_SQLiIntruderTest/MessageBody: harness.go:115: String Value: # from wapiti
    Test_Say_SQLiIntruderTest/MessageBody: harness.go:115: String Value: sleep(5)#
    Test_Say_SQLiIntruderTest/MessageBody: harness.go:115: String Value: 1 or sleep(5)#
...
--- PASS: Test_Say_SQLiIntruderTest (7.94s)
    --- PASS: Test_Say_SQLiIntruderTest/MessageBody (2.19s)
PASS
```

## Test Data Sets

Sample data sets can be found in [/payloads/xxx/testdata.json](https://github.com/mercari/testdeck/payloads/) where xxx is the payload type. All payload txt files are copied from [swisskyrepo/PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings).

Types of data sets:

- Input Validation: The test case will fail if the expected error message was not returned. In addition to string input, int, float, and boolean are also supported.

```
"string": [
    {
      "files": [
        "../payloads/input_validation/strings.txt"
      ],
      "type": "input validation",
      "expected": {
        "errorMessage": ""
      }
    }
  ],
```

- SQL Injection: Since it is difficult to test for SQLi automatically, only timing will be used. All payloads attempt to sleep more than 1s so the test will fail if the response time was greater.

```
  "string": [
    {
      "files": [
        "../payloads/sql_injection/Generic_TimeBased.txt"
      ],
      "type": "sql injection",
      "expected": {
        "timeDelay": 1
      }
    }
  ]
```

- XSS: Only reflected XSS can be tested at the moment. By leaving the error message blank, the test will fail only if the malicious payload was found anywhere within the response.

```
"string": [
    {
      "files": [
        "../payloads/xss/IntrudersXSS.txt",
        "../payloads/xss/XSS_Polyglots.txt",
        "../payloads/xss/XSSDetection.txt"
      ],
      "type": "reflected xss",
      "expected": {
        "errorMessage": ""
      }
    }
  ]
```

## Limitations

Currently, the intruder can only inject values into string, int, float, and boolean parameters in the request (i.e. it cannot access string/int/float/bool that are inside structs). We are hoping to support this functionality in the future.