package intruder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseInputValidationJson(t *testing.T) {
	data, err := ParseInputValidationTestDataFromJson("../payloads/input_validation/testdata.json")
	if err != nil {
		t.Fatalf("Failed to parse input validation testdata data from json file, got %s", err.Error())
	}

	assert.NotNil(t, data)
}

func Test_ReadStringsFromTextFile(t *testing.T) {
	strings, err := GetStringArrayFromTextFile("../payloads/input_validation/strings.txt")
	if err != nil {
		t.Fatalf("Failed to read from text file, got %s", err.Error())
	}

	assert.NotNil(t, strings, "Failed to create string array from text file")
}

func Test_ReadIntsFromTextFile(t *testing.T) {
	ints, err := GetIntArrayFromTextFile("../payloads/input_validation/integers.txt")
	if err != nil {
		t.Fatalf("Failed to read from text file, got %s", err.Error())
	}

	assert.NotNil(t, ints, "Failed to create int array from text file")
}

func Test_ReadFloatsFromTextFile(t *testing.T) {
	ints, err := GetFloatArrayFromTextFile("../payloads/input_validation/floats.txt")
	if err != nil {
		t.Fatalf("Failed to read from text file, got %s", err.Error())
	}

	assert.NotNil(t, ints, "Failed to create float array from text file")
}

func Test_ReadBoolsFromTextFile(t *testing.T) {
	ints, err := GetBoolArrayFromTextFile("../payloads/input_validation/booleans.txt")
	if err != nil {
		t.Fatalf("Failed to read from text file, got %s", err.Error())
	}

	assert.NotNil(t, ints, "Failed to create bool array from text file")
}
