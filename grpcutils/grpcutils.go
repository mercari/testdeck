package grpc

import (
	"context"
	"fmt"
	"reflect"
)

/*
grpc_helper.go: Helper methods for testing gRPC endpoints
*/

// Calls a grpc method by its name. Returns the response in generic form.
// client: The grpc client to use
// methodName: The name of the method to call
// req: The request casted to generic interface{}
// returns the response casted to generic interface{} and error
func CallRpcMethod(ctx context.Context, client interface{}, methodName string, req interface{}) (interface{}, error) {
	var err error
	m := reflect.ValueOf(client).MethodByName(methodName)
	in := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
	res := m.Call(in)

	// most rpc methods return two items (response and error) but check the length in case
	if len(res) != 2 {
		return nil, fmt.Errorf("expected 2 return items but got %d", len(res))
	}

	if res[1].Interface() != nil {
		err = res[1].Interface().(error)
	}

	return res[0].Interface(), err
}
