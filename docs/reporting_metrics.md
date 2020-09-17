# Reporting and Metrics

As with any form of test automation, it is recommended to store metrics such as tests run, duration, results, etc. for statistical analysis and reporting.

To serve as an example, we have included our test results DB schema below and some sample code so that you can build your own reporting and metrics collection system.

NOTE: Testdeck can be used without the reporting and metrics system. If you do not set the environment variable `DB_URL` , then the tests will run without saving results to anywhere. Results will only be saved to the Kubernetes pod's logs, so you can use the following command to see the test results: `kubectl logs <your-pod-name>`

![Testdeck Test Results DB Schema](images/metrics.png?raw=true)

Relevant Files:
- service
    - db
        - db.go

Once test results are saved to the DB, you can create your own reporting dashboard or integrate another dashboard tool to read and display the results.

![Testdeck Test Results Dashboard](images/reporting.png?raw=true)

## Debugging

### If test results are saved to a DB:

Afer a test is run, Testdeck saves the results and statistics of the test case to the DB so that we can easily see at which step a the test case failed at. E2E tests can take longer than unit tests due to the need to perform multiple setup steps, so this allows us to narrow down possible reasons for failure:

- If a test case failed at Arrange or After: This usually means that something your test depends on (e.g. test account and test data) is broken.
- If a test case failed at Act: This usually means that there is something wrong with your service (e.g. the endpoint returned an error, the service could not be reached, etc.).
- If a test case failed at Assert: This usually means that the response returned is different from what was originally expected. Perhaps the response format or spec changed so the test case needs to be updated (or you just found a bug).

### If test results are NOT saved to a DB:

The Kubernetes pod that the tests are executed on will save the results as logs. To see the logs, use the command: `kubectl logs <your-pod-name>`

