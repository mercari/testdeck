# Security Testing Guide

## Why Should We Use Testdeck for Automating Security Tests?

![QA vs Security Testing](images/sec.png?raw=true)

If you think about it, QA testing and Security testing are actually quite similar, like two sides of the same coin.

QA tests from the perspective of an average or slightly curious user. Therefore, they will test "Happy Path" cases, input validation (edge cases and strange input), common/accidental error cases, and business logic.

Security tests from the perspective of a malicious user. Therefore, they will test input validation (malicious payloads and fuzz testing), authentication and authorization, intentional/exploitative error cases, and business logic.

There is a natural overlap in their testing so it makes sense to use the same tool to automate both types of testing.

## Types of Test Cases

### Input Validation

For each endpoint and input parameter, you should test the following:

#### Boundary input values
- Invalid input values (null values, non-numeric characters, invalid string format, strange ASCII characters, 0, negative numbers, decimal/fraction numbers, etc.)
- Extreme input values (a string of 40 million characters, infinity number, etc.)
- Malicious strings often used in SQL Injection, XSS, Command Injection, and other common input-related attacks

#### Authentication and Authorization
These are the tests that verify user permission and token logic.

##### Tokens
For each endpoint, you may want to test the following:

- Request with an authorized token
- Request with no token
- Request with an invalid token
- Request with an expired token
- Request with incorrect user type (user lacking authorization)
- Request with incorrect token scope (attempting to execute an unauthorized action)
- Request with the token of another user (session hijacking)

#### Intentional and Exploitative Error Cases

These are tests that verify proper error handling. The test cases will be different depending on each service but below are some ideas of what to test for:

- Verify that error messages do not accidentally reveal information about the system that can potentially be useful to attackers
    - e.g. A login error should return Incorrect email or password instead of specifying which field was incorrect because doing so helps the attacker brute force into the account
    - e.g. Trying to get a user profile with a non-existent user ID should return a generic error instead of specifying that the user does not exist because doing so helps the attack enumerate a list of valid user IDs
- Verify that extreme input values and other unexpected client behavior is properly handled and does not cause internal server errors, high memory usage, high response latency, or other undesirable behavior that can impact the rest of the traffic.

#### Configuration Flaws

These are tests that check for malformed requests and malicious header settings. These apply only to HTTP requests. Below are some ideas of what to test for:

- Verify that only GET, POST, HEAD, and OPTIONS methods are accepted
- Verify that the correct response type is returned regardless of the value specified in the request's Accept header
- Verify that response Content-Type header is set to an accepted value
- Verify that software version information is not disclosed anywhere in the response body or headers
- Verify that Strict-Transport-Security header is enforced
- Verify that if Origin header is changed, the response's Access-Control-Allow-Origin header is also changed
- Verify that the Authorization header is required

#### Business Logic
These are tests that verify for intended business logic and behavior. These test cases are closer to E2E error cases because they verify the user flow and error handling of strange user behavior. Below are some ideas of what to test:

- Can you do the steps of the user flow in a different order than intended? (e.g. Can you give your buyer a rating before you even shipped the item?)
- Can you skip a step in the user flow? (e.g. Can you purchase an item without registering a payment method?)