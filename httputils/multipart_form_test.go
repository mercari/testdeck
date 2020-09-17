package httputils

import (
	"fmt"
	"mime"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DemoStruct struct {
	Key   string
	Value string
}

func (d *DemoStruct) WriteField(writer *multipart.Writer) error {
	writer.WriteField(d.Key, d.Value)
	return nil
}

type DemoStringer string

func (d DemoStringer) String() string {
	return string(d)
}

func Test_Multipart(t *testing.T) {
	in := struct {
		ID            string       `multipart:"field"`
		StringParam   string       `json:"string_param" multipart:"field"`
		IntParam      int          `json:"int_param" multipart:"field"`
		UInt64Param   uint64       `json:"uint64_param" multipart:"field"`
		Float64Param  float64      `json:"float64_param" multipart:"field"`
		StringerParam fmt.Stringer `json:"stringer_param" multipart:"field" optional:"false"`
		StructParam   *DemoStruct  `json:"-" multipart:"custom"`
	}{
		ID:            "id_here",
		StringParam:   "string param here",
		IntParam:      -1,
		UInt64Param:   42,
		Float64Param:  3.14,
		StringerParam: DemoStringer("demo stringer value"),
		StructParam: &DemoStruct{
			Key:   "demo_struct_key",
			Value: "demo struct value",
		},
	}

	body, contentType, err := CreateMultipartBody(&in)
	if err != nil {
		t.Fatalf("err should be nil. got: %v", err)
	}
	if body == nil {
		t.Errorf("http should not be nil")
	}
	if contentType == "" {
		t.Errorf("content type should be set")
	}

	_, params, _ := mime.ParseMediaType(contentType)
	reader := multipart.NewReader(body, params["boundary"])
	maxMemory := int64(1024 * 1024)
	form, err := reader.ReadForm(maxMemory)
	if err != nil {
		t.Fatalf("Failed to read form. %v", err)
	}

	assert.Equal(t, in.ID, form.Value["ID"][0])
	assert.Equal(t, in.StringParam, form.Value["string_param"][0])
	assert.Equal(t, fmt.Sprint(in.IntParam), form.Value["int_param"][0])
	assert.Equal(t, fmt.Sprint(in.UInt64Param), form.Value["uint64_param"][0])
	assert.Equal(t, fmt.Sprint(in.Float64Param), form.Value["float64_param"][0])
	assert.Equal(t, in.StringerParam.String(), form.Value["stringer_param"][0])
	assert.Equal(t, in.StructParam.Value, form.Value[in.StructParam.Key][0])
}

func Test_MultipartOptional(t *testing.T) {
	in := struct {
		StringParam        string `json:"string_param" multipart:"field"`
		OptionalSetParam   string `json:"set_param" multipart:"field" optional:"true"`
		OptionalUnsetParam string `json:"unset_param" multipart:"field" optional:"true"`
	}{
		StringParam:      "string param here",
		OptionalSetParam: "set parameter value",
	}

	body, contentType, err := CreateMultipartBody(&in)
	if err != nil {
		t.Fatalf("err should be nil. got: %v", err)
	}
	if body == nil {
		t.Errorf("http should not be nil")
	}
	if contentType == "" {
		t.Errorf("content type should be set")
	}

	_, params, _ := mime.ParseMediaType(contentType)
	reader := multipart.NewReader(body, params["boundary"])
	maxMemory := int64(1024 * 1024)
	form, err := reader.ReadForm(maxMemory)
	if err != nil {
		t.Fatalf("Failed to read form. %v", err)
	}

	assert.Equal(t, in.StringParam, form.Value["string_param"][0])
	assert.Equal(t, in.OptionalSetParam, form.Value["set_param"][0])
	val, ok := form.Value["unset_param"]
	if ok {
		t.Errorf("unset_param was set when it should not be, got value: %v", val)
	}
}

func Test_MultipartMultipleCustomFields(t *testing.T) {
	in := struct {
		StructParamA   *DemoStruct `multipart:"custom"`
		StructParamB   *DemoStruct `multipart:"custom"`
		StructParamNil *DemoStruct `json:"nil_struct_key" multipart:"custom"`
	}{
		StructParamA: &DemoStruct{
			Key:   "a_struct_key",
			Value: "A struct value",
		},
		StructParamB: &DemoStruct{
			Key:   "b_struct_key",
			Value: "B struct value",
		},
	}

	body, contentType, err := CreateMultipartBody(&in)
	if err != nil {
		t.Fatalf("err should be nil. got: %v", err)
	}
	if body == nil {
		t.Errorf("http should not be nil")
	}
	if contentType == "" {
		t.Errorf("content type should be set")
	}

	_, params, _ := mime.ParseMediaType(contentType)
	reader := multipart.NewReader(body, params["boundary"])
	maxMemory := int64(1024 * 1024)
	form, err := reader.ReadForm(maxMemory)
	if err != nil {
		t.Fatalf("Failed to read form. %v", err)
	}

	assert.Equal(t, in.StructParamA.Value, form.Value[in.StructParamA.Key][0])
	assert.Equal(t, in.StructParamB.Value, form.Value[in.StructParamB.Key][0])
	val, ok := form.Value["nil_struct_key"]
	if ok {
		t.Errorf("StructParamNil was set when it should not be, got value: %v", val)
	}
}
