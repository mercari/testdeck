package fuzzer

import (
	"context"
	"fmt"
	"github.com/google/gofuzz"
	"github.com/mercari/testdeck/grpcutils"
	"reflect"
	"strings"
	"testing"
)

/*
fuzzer.go: Helper methods for executing fuzz tests on an endpoint

Currently, Google's GoFuzz (https://github.com/google/gofuzz) is being used to generate random values
If this idea is implemented by the Golang team, we should change to using it instead: https://go.googlesource.com/proposal/+/master/design/draft-fuzzing.md

*/

const DefaultFuzzRounds = 10000
const DefaultNilChance = 0.05

// Represents configurable options for fuzzing
type FuzzOptions struct {
	rounds    int      // the number of inputs to try
	nilChance float64  // the probability of getting a nil value
	debugMode bool     // prints the values tried (for debugging purpose)
	ignoreNil []string // the names of fields that do not support empty/nil values (the fuzzer will not try an empty value when fuzzing these fields)
}

// Runs a fuzz test on all fields of the specified GRPC endpoint
// client is the client for the microservice (e.g. echoClient)
// req is a sample, valid request (the fuzzer will mutate the values in this request)
// opts are configurations for the fuzzing, if not included the default settings will be used
func FuzzGrpcEndpoint(t *testing.T, ctx context.Context, client interface{}, methodName string, req interface{}, opts ...FuzzOptions) {

	// get parameters of the sample request using reflection because we do not know the protobuf type
	fieldNames := reflect.TypeOf(req).Elem()
	fieldValues := reflect.ValueOf(req).Elem()

	// loop through each parameter of the endpoint
	for i := 0; i < fieldValues.NumField(); i++ {
		fieldName := fieldNames.Field(i).Name

		// skip this parameter if it is an automatically-generated field (field name starts with XXX)
		if strings.HasPrefix(fieldName, "XXX") {
			break
		}

		// run fuzz tests on this field
		if len(opts) > 0 {
			FuzzThisField(t, ctx, client, methodName, req, fieldName, opts[0])
		} else {
			FuzzThisField(t, ctx, client, methodName, req, fieldName)
		}
	}
}

// Runs the specified field of the endpoint
// client is the client for the microservice (e.g. echoClient)
// req is a sample, valid request (the fuzzer will mutate the values in this request)
// fieldName is the field to fuzz
// opts are configurations for the fuzzing, if not included the default settings will be used
func FuzzThisField(t *testing.T, ctx context.Context, client interface{}, methodName string, req interface{}, fieldName string, opts ...FuzzOptions) {

	// the log of values that were tried; it is printed only if opts.debugMode is set to true
	var log []string

	// default fuzz settings
	var rounds int = DefaultFuzzRounds
	var nilChance float64 = DefaultNilChance
	var debugMode bool = false

	// change fuzz settings if config was passed in
	if len(opts) > 0 {
		rounds = opts[0].rounds

		// if field does not support empty/nil values, set probability of nil values to 0
		if contains(fieldName, opts[0].ignoreNil) {
			nilChance = 0
		} else {
			nilChance = opts[0].nilChance
		}

		debugMode = opts[0].debugMode
	}

	// get the current field to fuzz
	field := reflect.ValueOf(req).Elem().FieldByName(fieldName)

	// fuzz with different input depending on the type of the field
	switch field.Kind() {

	// String type
	case reflect.String:
		// make a copy of the normal value of this field
		copy := field.String()

		for i := 0; i < rounds; i++ {
			// fuzz string
			var s string
			fuzz.New().NilChance(nilChance).Fuzz(&s)
			field.SetString(s)
			_, err := grpc.CallRpcMethod(ctx, client, methodName, req)
			if err != nil {
				t.Errorf("[FAIL] Fuzzing %s > %s: \"%s\" --> ERROR: %s\n", methodName, fieldName, s, err.Error())
			}

			// add input to debug log
			log = append(log, fmt.Sprintf("[PASS] Fuzzing %s > %s: \"%s\"\n", methodName, fieldName, s))
		}

		// set field back to the normal value when done fuzzing
		field.SetString(copy)

	case reflect.Int:
		// make a copy of the normal value of this field
		copy := field.Int()

		for i := 0; i < rounds; i++ {
			// fuzz int
			var j int
			fuzz.New().NilChance(nilChance).Fuzz(&j)
			field.SetInt(int64(j))
			_, err := grpc.CallRpcMethod(ctx, client, methodName, req)
			if err != nil {
				t.Errorf("Error occurred while fuzzing %s. Input: \"%d\"; Error: %s", methodName, j, err.Error())
			}

			// add input to debug log
			log = append(log, fmt.Sprintf("Fuzzing %s > %s: \"%d\"\n", methodName, fieldName, j))
		}

		// set field back to the normal value when done fuzzing
		field.SetInt(copy)

	case reflect.Float64:
		// make a copy of the normal value of this field
		copy := field.Float()

		for i := 0; i < rounds; i++ {
			// fuzz float
			var f float64
			fuzz.New().NilChance(nilChance).Fuzz(&f)
			field.SetFloat(f)
			_, err := grpc.CallRpcMethod(ctx, client, methodName, req)
			if err != nil {
				t.Errorf("Error occurred while fuzzing %s. Input: \"%b\"; Error: %s", methodName, f, err.Error())
			}

			// add input to debug log
			log = append(log, fmt.Sprintf("Fuzzing %s > %s: \"%b\"\n", methodName, fieldName, f))
		}

		// set field back to the normal value when done fuzzing
		field.SetFloat(copy)
	}

	// print log if debug mode
	if debugMode {
		fmt.Println(log)
	}
}

// Helper function to determine if an array of strings contains a certain string
func contains(s string, array []string) bool {
	for _, b := range array {
		if b == s {
			return true
		}
	}
	return false
}
