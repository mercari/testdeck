# Utility Methods

## grpcutils

Call a grpc method by passing the following parameters into `CallRpcMethod`:
- context
- grpc client of the service
- name of the method to call
- the request body

```
var req = &pb.SayRequest{MessageId: "test", MessageBody: "test"}

res, err := grpc.CallRpcMethod(context.TODO(), echoClient, "Say", req)
```


## httputils

Features:
- send HTTP requests
- build a multipart form from a struct
- connect a debugging proxy such as Charles or Burpsuite to intercept traffic for analysis, modification, replay, etc.

Send a HTTP request by passing the following parameters into `SendHTTPRequest()`:
- the HTTP method (POST, GET, etc.)
- URL
- json body of the request

```
var headers = map[string]string{
	"Authorization": apiClient.MatToken,
	"Content-Type":  constants.JsonContentType,
}

var body, _ = json.Marshal(map[string]string{
	"message_id":   "test",
	"message_body": "test",
})

res, _, err = tdhttp.SendHTTPRequest(http.MethodPost, httpUrl, bytes.NewBuffer(body), headers)
if err != nil {
	t.Fatalf("An unexpected failure occurred: %s", err.Error())
}
```

For examples of how to make a multipart form from a struct (to use in HTTP requests), please see the httputils unit tests.

To connect a debugging proxy such as Charles or Burpsuite, simply add the following line of code to the beginning of your test case or to the testing main method:

```
ConnectToProxy("http://<your-ip-here>:<your-port-here>")
```