package httputils

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"reflect"

	"github.com/pkg/errors"
)

/*
multipart_form.go: Contains helper methods to convert a type struct into a multipart form
*/

type MultipartFieldWriter interface {
	WriteField(*multipart.Writer) error
}

const (
	jsonTag      = "json"
	multipartTag = "multipart"
	optionalTag  = "optional"
)

type FieldProperties struct {
	MultipartType string
	Value         string
	Optional      bool
	IsZero        bool
	Ref           MultipartFieldWriter
}

/*
Converts a type struct into a multipart form for use in HTTP requests
*/
func CreateMultipartBody(it interface{}) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// for each field in the struct, write it to a multipart form
	fields := structFieldsMap(it)
	for key, val := range fields {
		switch val.MultipartType {
		case "field":
			if !val.Optional || !val.IsZero {
				writer.WriteField(key, val.Value)
			}
		case "custom":
			if val.Ref != nil {
				err := val.Ref.WriteField(writer)
				if err != nil {
					return nil, "", errors.Wrap(err, "Failed during custom multipart field write.")
				}
			}
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to write multipart/form http.")
	}

	return body, writer.FormDataContentType(), nil
}

/*
Helper method to convert type struct data into FieldProperties type
*/
func structFieldsMap(it interface{}) map[string]FieldProperties {
	fmap := make(map[string]FieldProperties)
	val := reflect.ValueOf(it).Elem()
	multipartFieldWriterType := reflect.TypeOf((*MultipartFieldWriter)(nil)).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valField := val.Field(i)
		fp := FieldProperties{}

		// property: MultipartType
		fp.MultipartType, _ = typeField.Tag.Lookup(multipartTag)

		// property: Optional
		optionalTag, _ := typeField.Tag.Lookup(optionalTag)
		fp.Optional = optionalTag == "true"

		// property: IsZero
		// property: Value
		if fp.MultipartType == "field" {
			fp.IsZero = isZero(valField)
			fp.Value = fmt.Sprint(valField.Interface())
		}

		// property: Ref
		if typeField.Type.Implements(multipartFieldWriterType) && !valField.IsNil() {
			fp.Ref = valField.Interface().(MultipartFieldWriter)
		}

		itemKey := typeField.Name
		jsonKey, jsonTagAvailable := typeField.Tag.Lookup(jsonTag)
		if jsonTagAvailable {
			itemKey = jsonKey
		}
		fmap[itemKey] = fp
	}
	return fmap
}

// Modified version of https://stackoverflow.com/a/23555352
// This version avoids recursion by not supporting Array and Struct.
// If is a struct or array, this will panic.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array, reflect.Struct:
		panic("isZero: Array and Struct is not supported!")
	default:
		// concrete types and interface
		z := reflect.Zero(v.Type())
		return v.Interface() == z.Interface()
	}
}
