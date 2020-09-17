# Setup

1. Clone (or fork) the repository. As with any automation framework, you may need to modify the framework to suit your product’s specific needs so we suggest that you clone (or fork if you intend to contribute) the repository and modify it as needed.

2. (If you would like to save test results to a DB for visualization and statistical analysis) Set up your DB and modify db.go to fit your DB schema

3. Write your test cases: following the documentation and style guide sample code, create a suite of test cases.

4. Create a main method to run your tests. This method uses Testdeck to start a test service to run all of your tests in:

```
import "github.com/mercari/testdeck/service"
func TestMain(m *testing.M) {
	service.Start(m)
}
```

5. In your Dockerfile, add steps to package your tests into a `.test` binary file and run them on container start. For example:

```
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go test -c my_project/testdeck_tests/ -o /go/bin/testdeck.test // where testdeck_tests is the directory in your project containing all Testdeck test cases

COPY --from=0 /go/bin/testdeck.test /bin/testdeck.test
CMD ["/bin/testdeck.test", "-test.v", "-test.parallel=x"] // where x is the number of parallel tests to run
```

6. Create a manifest to deploy the image that was created from the Dockerfile in the step above. Below is a sample manifest:

```
apiVersion: batch/v1
kind: Job
metadata:
  name: testdeck
  namespace: your-namespace-here
spec:
  template:
    spec:
      containers:
        - env:
            - name: ENV
              value: development
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: JOB_NAME
              valueFrom:
                fieldRef:
                  fieldPath: 'metadata.labels[''job-name'']'
            - name: RUN_AS
              value: job
            - name: GCP_PROJECT_ID
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: ASSET_PATH
              value: /assets
          image: <your Docker image link here>
          imagePullPolicy: Always
          name: testdeck
      restartPolicy: Never

      ...
      <other information omitted>
```

7. Configure your Spinnaker pipeline to delete any existing Testdeck pods first, and then deploy using the manifest above.

8. After deployment, you should be able to see your test job running as a pod! Use the following command to check the status of the test run: `kubectl get pods`