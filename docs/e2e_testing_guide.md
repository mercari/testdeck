# E2E/Integration Testing Guide

## Using the Testdeck Lifecycle

Testdeck test cases are divided into four "lifecycle" stages. You do not have to use all stages; you can just use the ones that you need (i.e. if you do not require any setup, there is no need to declare Arrange in your test case).

- Arrange: This is the setup stage. If you need to create test data, login to a test account, etc. you should put that code here.
- Act: This is the actual testing stage. Here, you should call the endpoint that you want to test.
- Assert: This is the verification stage. Here, you should add assertion statements to verify that the response returned in the Act stage matches what you expect.
- After: This is the cleanup stage. If you to do anything after the test (e.g. restore data back to original state, delete data, etc.) you should put that code here.

![Testdeck Lifecycle Stages](images/lifecycle.png?raw=true)

## Debugging Failed Test Cases

Please see the [Reporting and Metrics](../../docs/reporting.md) doc for more tips on how to debug.

## Types of Test Cases

### Testing "Happy Path"
These are the success cases- scenarios where the user's behavior is expected and normal. All services should cover this type of test for all endpoints.

### Testing Error Cases
These are the failure cases- scenarios where the user does something strange and unexpected, so a handled error should return. It is recommended that you cover this type of test for all possible errors.
