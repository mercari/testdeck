package intruder

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

/*
testdata_helper.go: Various helper methods for generating testdata data, etc.
*/

// ----------
// structs representing json testdata data sets
// ----------

// data for testing input validation
// test data need to be separated by data type so that the fuzzer can feed the proper data type into the parameter
type InputValidationTestData struct {
	Strings []JsonDataSet `json:"string"`
	Ints    []JsonDataSet `json:"int"`
	Floats  []JsonDataSet `json:"float"`
	Bools   []JsonDataSet `json:"bool"`
}

// represents a json data set
// Files is the list of intruder .txt files to use
// Expected is the expected result
type JsonDataSet struct {
	Files    []string       `json:"files"`
	Type     string         `json:"type"`
	Expected ExpectedResult `json:"expected"`
}

type ExpectedResult struct {
	// TODO: Clarify what else needs to be checked in the response
	ErrorMessage string `json:"errorMessage"`
	TimeDelay    int    `json:"timeDelay"`
}

// Parse input validation json testdata data into a struct
func ParseInputValidationTestDataFromJson(file string) (InputValidationTestData, error) {
	jsonFile, _ := ioutil.ReadFile(file)
	var data InputValidationTestData
	err := json.Unmarshal(jsonFile, &data)
	return data, err
}

// Parses an intruder .txt file into a string array
func GetStringArrayFromTextFile(filename string) ([]string, error) {
	var array []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		array = append(array, scanner.Text())
	}

	return array, nil
}

// Parses an intruder .txt file into an int array
func GetIntArrayFromTextFile(filename string) ([]int, error) {
	var array []int
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, err
		}
		array = append(array, i)
	}

	return array, nil
}

// Parses an intruder .txt file into a float array
func GetFloatArrayFromTextFile(filename string) ([]float64, error) {
	var array []float64
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return nil, err
		}
		array = append(array, i)
	}

	return array, nil
}

// Parses an intruder .txt file into a bool array
func GetBoolArrayFromTextFile(filename string) ([]bool, error) {
	var array []bool
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i, err := strconv.ParseBool(scanner.Text())
		if err != nil {
			return nil, err
		}
		array = append(array, i)
	}

	return array, nil
}
