# Testdeck

Testdeck is a framework for integration, end-to-end (E2E), and security testing of gRPC microservices written in Golang.

Please see the [docs](https://github.com/mercari/testdeck/docs) folder for documentation and tutorials.

# Features

- Integration/E2E testing for gRPC microservices written in Golang
- Fuzz testing of gRPC endpoints (using [google/gofuzz](https://github.com/google/gofuzz))
- Injection of malicious payloads (from [swisskyrepo/PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings)), similar to Burpsuite's Intruder function
- Utility methods for gRPC/HTTP requests
- Connecting a debugging proxy such as Charles or Burpsuite to analyze, modify, replay, etc. requests

# How to Use

As with all test automation frameworks, you will most likely need to do some customization and fine-tuning to tailor this tool for your organization. It is recommended that you clone this repository and build on top of it to suit your needs.

To learn how to build your own test automation system using Testdeck, please see the [setup guide](https://github.com/mercari/testdeck/docs/setup.md) and the blog article.

# Contribution

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

# License

Copyright 2020 Mercari, Inc.

Licensed under the MIT License.
