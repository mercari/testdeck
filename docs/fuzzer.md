# Fuzzer

The Fuzzer uses [google/gofuzz](https://github.com/google/gofuzz) to generate random data to feed into endpoints.

Relevant files:

- fuzzer
    - fuzzer_test.go
    - fuzzer.go

## How to Use

Pass the following parameters into `FuzzGrpcEndpoint()`:

- t (the current test case)
- context
- grpc client of the service
- the rpc method to test
- a valid, sample request (this will act as the base for the fuzzer)
- FuzzOptions (optional, explained below)

```
var sampleRequest = &pb.SayRequest{MessageId: "test", MessageBody: "test"}

func Test_Gofuzz(t *testing.T) {

	// insert your setup steps here, including getting the service client

	test.Act = func(td *testdeck.TD) {
		FuzzGrpcEndpoint(t, context.TODO(), client, "Say", sampleRequest, FuzzOptions{rounds: 100, nilChance: 0.01, debugMode: true, ignoreNil: []string{"MessageId"}})
	}

	test.Run(t, t.Name())
}
```


The FuzzOptions struct passed as an optional parameter contains configuration for the fuzzer. If it is not passed in, default configuration will be used.

```
// Represents configurable options for fuzzing
type FuzzOptions struct {
	rounds int // the number of inputs to try
	nilChance float64 // the probability of getting a nil value
	debugMode bool // prints the values tried (for debugging purpose)
	ignoreNil []string // the names of fields that do not support empty/nil values (the fuzzer will not try an empty value when fuzzing these fields)
}
```

The output will look similar to below:

```
=== CONT  Test_Gofuzz/Test_Gofuzz
Test_Gofuzz: fuzzer.go:107: [FAIL] Fuzzing Say > MessageId: "" --> ERROR: rpc error: code = InvalidArgument desc = failed to validate request: invalid SayRequest.MessageId: value length must be between 1 and 64 runes, inclusive
[[PASS] Fuzzing Say > MessageId: "鯛Ȁ"
 [PASS] Fuzzing Say > MessageId: "uǷM鈫"
 [PASS] Fuzzing Say > MessageId: "u們Ĭ驂H嫹w"
 ...
 ```

## Limitations

Currently, the fuzzer can only fuzz string parameters in the request (i.e. it cannot access strings inside structs). We are hoping to support this functionality in the future.