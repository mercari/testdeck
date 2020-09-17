package intruder

import (
	"context"
	"fmt"
	"github.com/mercari/testdeck"
	"github.com/mercari/testdeck/grpcutils"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

// A helper function that verifies that the response matches the expected results fetched from the json test data file
func VerifyIntruderTestResults(t *testdeck.TD, data JsonDataSet, res interface{}, duration time.Duration, input string, err error) {

	switch data.Type {
	case "input validation":
		if data.Expected.ErrorMessage != "" {
			// verify that an error was returned
			assert.NotNil(t, err, "FAIL: Error was not returned as expected")
			if err != nil {
				// if an error was returned, verify that the error message is correct
				assert.Contains(t, err.Error(), data.Expected.ErrorMessage, "FAIL: Error message is different from expected")
			}
		} else { // success case checks
			assert.NotNil(t, res, "FAIL: Response is nil")
			assert.Nil(t, err, "FAIL: Unexpected error was returned")
		}
	case "sql injection":
		assert.False(t, duration.Seconds() > float64(data.Expected.TimeDelay), "WARNING: Potential SQLi found")
	case "reflected xss":
		str := fmt.Sprintf("%v", res)
		assert.NotContains(t, str, input, "WARNING: Potential reflected XSS found")
	}
}

// Runs a fuzz test on all parameters of this request
// req is a sample request struct specified in the protobuf file (e.g. pb.SayRequest)
// function is the function to be called
// dataFile is the json file where fuzzing data will come from
func RunIntruderTests(t *testing.T, ctx context.Context, td testdeck.TestCase, client interface{}, methodName string, req interface{}, data InputValidationTestData) {

	// get parameters of the sample request using reflection because we do not know the protobuf type
	fieldNames := reflect.TypeOf(req).Elem()
	fieldValues := reflect.ValueOf(req).Elem()

	// for each parameter field in this endpoint
	for i := 0; i < fieldValues.NumField(); i++ {
		fieldName := fieldNames.Field(i).Name

		// skip this parameter if it is an automatically-generated field (field name starts with XXX)
		if strings.HasPrefix(fieldName, "XXX") {
			break
		}

		// run fuzz tests on this field
		TestThisField(t, ctx, td, client, methodName, req, fieldName, data)
	}
}

// This method generates an actual testdeck test case to fuzz the specified field
// req is the sample request struct
// fieldName is the current field to fuzz
// function is the fuzzing function
// dataFile is the json file where fuzzing data will come from
func TestThisField(t *testing.T, ctx context.Context, tc testdeck.TestCase, client interface{}, methodName string, req interface{}, fieldName string, testDataSet InputValidationTestData) {

	var (
		err error
		res interface{}
	)
	// Act
	tc.Act = func(t *testdeck.TD) {
		// get the current field to fuzz
		field := reflect.ValueOf(req).Elem().FieldByName(fieldName)

		// fuzz with different input depending on the type of the field
		switch field.Kind() {

		// String type
		case reflect.String:
			// make a copy of the normal value of this field
			copy := field.String()

			for _, set := range testDataSet.Strings {
				// loop through the intruder .txt files specified in the json file
				for _, file := range set.Files {
					strings, _ := GetStringArrayFromTextFile(file)
					// loop through all the strings in the intruder .txt file
					for _, s := range strings {
						t.Logf("String Value: %v", s)
						field.SetString(s)
						start := time.Now()
						res, err = grpc.CallRpcMethod(ctx, client, methodName, req)
						duration := time.Since(start)
						VerifyIntruderTestResults(t, set, res, duration, s, err)
					}
				}

				// reset parameter back to the normal value before fuzzing the next field
				field.SetString(copy)
			}

			// Int type
		case reflect.Int:
			// make a copy of the normal value of this field
			copy := field.Int()

			for _, set := range testDataSet.Ints {
				// loop through the intruder .txt files specified in the json file
				for _, file := range set.Files {
					ints, _ := GetIntArrayFromTextFile(file)
					// loop through all the ints in the intruder .txt file
					for _, i := range ints {
						t.Logf("Int Value: %v", i)
						field.SetInt(int64(i))
						res, err = grpc.CallRpcMethod(ctx, client, methodName, req)
						VerifyIntruderTestResults(t, set, res, 0, "", err)
					}
				}
				// reset parameter back to the normal value before fuzzing the next field
				field.SetInt(copy)
			}

			// Float type
		case reflect.Float64:
			// make a copy of the normal value of this field
			copy := field.Float()

			for _, set := range testDataSet.Floats {
				// loop through all the intruder .txt files specified in the json file
				for _, file := range set.Files {
					floats, _ := GetFloatArrayFromTextFile(file)
					// loop through all the floats in the intruder .txt file
					for _, f := range floats {
						t.Logf("Float Value: %v", f)
						field.SetFloat(f)
						//res, err = CallFunction(function, req, apiClient)
						res, err = grpc.CallRpcMethod(ctx, client, methodName, req)
						VerifyIntruderTestResults(t, set, res, 0, "", err)
					}
				}
				// reset parameter back to a normal value before fuzzing the next field
				field.SetFloat(copy)
			}

			// Bool type
		case reflect.Bool:
			// make a copy of the normal value of this field
			copy := field.Bool()

			for _, set := range testDataSet.Bools {
				// loop through all the strings in the intruder .txt file
				for _, file := range set.Files {
					bools, _ := GetBoolArrayFromTextFile(file)
					// loop through all the bools in the intruder .txt file
					for _, b := range bools {
						t.Logf("Bool Value: %v", b)
						field.SetBool(b)
						//res, err = CallFunction(function, req, apiClient)
						res, err = grpc.CallRpcMethod(ctx, client, methodName, req)
						VerifyIntruderTestResults(t, set, res, 0, "", err)
					}
				}
				// reset parameter back to a normal value before fuzzing the next field
				field.SetBool(copy)
			}
		}
	}

	tc.Run(t, fieldName)
}
