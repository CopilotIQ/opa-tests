# Open Policy Agent (OPA) Integration Tests Framework

## Releasing

- Update the `version` in the `Makefile` through commit to `main`
- Merge from `main` to `release`
- For depedent projects, download the binary for the new version into the source dir (e.g. `out/bin/opatest`)

### Copyright & Licensing

**The code is copyright (c) 2023 CopilotIQ Inc. All rights reserved**

_This code is licensed under the terms of the Apache License 2.0, see LICENSE for more details_

# Motivation

This repository contains a policy validation framework for [Open Policy Agent](https://opa.io) server policies.

Rego policies have a unit testing mechanism, but we found it somewhat cumbersome to use, and the overall structure is not particularly amenable to code re-use and structuring in modules.

This is a very simple Golang binary which will take a number of `Testcases` (described in YAML) and generate the appropriate JSON API request to send to an OPA server running in a container, and then assert the truthiness (or otherwise) of the returned response, against expectations.

The code in this repository is particularly amenable to be used with the [JWT-OPA](https://github.com/massenz/jwt-opa) OPA/Spring Security Integration library; however, it can be adapted to other request structures.

`TODO: this README is work in progress`

# Testing Rego Policies

## Testcase Generation

All Policies Tests are generated following the `Testcase` definitions (in YAML), and running the `opa-test` test generator and runner.

The Generator itself is written in Go and built using `make`, the binary is generated in `out/bin` (or can be downloaded from the `Releases` page).

## Testcases

A `Testcase` (a self-contained YAML descriptor<sup>1</sup>) describes a set of `Test`s which will be run in sequence; each `Test` will result in a JSON payload which, depending on the value of `expect` will be generated in either the `allow` or `deny` subfolder of `src/test/resouces` and will respectively expect a `{"result": true | false}` when the tests are run, from OPA.

Each OPA request carries along a JWT, which will be encoded and signed by the generator, and sent along with the request.

Each JWT carries a `sub` and an array or `roles` which will be used to authorize the request.<sup>2</sup>

A `policy request` essentially asks OPA to assert whether the given `role` is allowed to perform the given "action" (defined by the HTTP `method`) on the "entity" (encoded in the `path`).

In this respect a policy `P` defines a `mapping` from a set `{role, action, entity}` to `boolean`:

    P: {role, action, entity} -> boolean

Each `Test` in a `Testcase` then defines an invariant (or, if you will, an assertion) upon the whole set of policies `{P}` as it regards one specific mapping.

### Reference

**Note**

> This is still under active development and will change in future, especially as we add the ability to templatize the POST request (in JSON)

A `Testcase` currently defines the following fields:

```
testcase:
  name: Users
  description: "Policy tests for the /users API"
  iss: "example.issuer"

  target:
    policy: allow
    package: example

  tests:
    - name: "create_user"
      expect: false
      token:
        sub: "alice@gmail.com"
        roles:
          - USER
      resource:
        path: "/users"
        method: POST
    - ... more tests
```

- `name` and `description` are simply used for reporting purposes;
- `iss` will be inserted into the generated JWT, as the "issuer" of the Token;
- `target` defines the OPA `data` endpoint that will be used (see below);
- `tests` is a list of invocations to the OPA endpoint, which individually assert against a policy use case or condition (in this example, for example, that a `USER` role cannot create another user by `POST`ing to the `/users` API): see [Tests](#tests).

The `target` defines which policy will be invoked, in which Rego package, and ultimately it is composed in a way that is defined by OPA, for the `/v1/data` API; given the example above, we would `POST` the JSON requests (defined by each of the `tests`) to:

      http://<server>:<port>/v1/data/example/allow

If the Rego policy is in a `package com.example.users` and we wish to evaluate a `grant` rule, this would need to be described in the `Testcase` as follows:

```
  target:
    policy: grant
    package: com.example.users
```

and would be tested against the `/v1/data/com/example/users/grant` URL.

## Tests

A `test` is an assertion against a server's API (defined by the `resource` being accessed) by a given `subject` having a set of `roles` - the test asserts the value returned by the policy evaluation against the `expect` value:

```
    - name: "create_user"
      expect: false
      token:
        sub: "alice@gmail.com"
        roles:
          - USER
      resource:
        path: "/users"
        method: POST

```

will result in a JSON body to be sent to the test OPA server's `/v1/data/example/allow` endpoint, with a JWT with the following claims (amongst others):

```json
{
  "sub": "alice@gmail.com",
  "iss": "example.issuer",
  "roles": ["USER"],
  "expires_at": ...
}
```

the request would carry this JWT as an `api_token` field (base-64 encoded) and the following other fields:

```json
{
  "input": {
    "api_token": "eyJhbG...6I8eHnFU",
    "resource": {
      "host": "readings.dev.copilotiq.co",
      "path": "/users",
      "method": "POST"
    }
  }
}
```

This would succeed when the OPA server returns a response `{result: false}`, and fail with anything else (including an empty response, which indicates the required rule in the policy package does not exist).

**Note**

> To create or inspect JWTs you can use the [`jwtie`](https://github.com/massenz/jwtie) utility.

# Execute Tests

**NOTE: SOME PARTS OF THE BELOW ARE STILL BEING IMPLEMENTED**

Currently, `opatest` does not run the OPA container (see [`run-opa`](run-opa) for an example of how to do it) and will not generate the "bundle" from the policies (see [Bundle](#bundle) for an example of how to generate one).

Over the next few iterations, we will move all the functionality into the `opatest` binary; for now, to run the tests use:

`opa-test -manifest policies.json -opa http://localhost:8089 path/to/tests`

**END NOTE**

By default `opatest` assumes a certain directory structure for policies, and tests, but those defaults can be modified via command-line flags.

The structure resembles closes the one that Gradle enforces on projects, in a simplified form:

```
$(pwd)                    -- the current directory
  |
  -- src
  |    |
  |    -- main
  |    |   |
  |    |   -- rego         -- contains all *.rego OPA policies
  |    |   |
  |    |   -- resources    -- manifest.json
  |    |
  |    -- tests            -- contains all *.yaml Testcases
  |        |
  |        -- resources    -- contains all *.json Request Templates
  |
  -- out
      |
      -- reports           -- test results
```

The locations defined above can be changed using the following flags:

- `-manifest` path to `manifest.json`
- `-src` directory containing Rego (`*.rego`) policies (including subfolders)
- `-templates` directory for JSON Golang templates for the requests
- `-out` directory where the test results will be generated
- `path/to/tests` if present, the first argument will point to the folder containing the `*.yaml` testcases

All paths can be absolute or relative to the current folder.

`opatest -h` will provide more up-to-date details about flags and defaults.

Running `opatest` will cause the following to happen:

1. all Rego files will be "bundled" into a `tar.gz` archive stored in a temporary directory;

2. an OPA [TestContainer](https://testcontainers.io) will be launched, and the bundle loaded;

3. for each of the `Testcase` files in the `tests` directory, we will extract the list of `tests`;

4. for each one of them, we generate an encoded JWT, and the JSON request body, then POST it to the test OPA;

5. the `result` returned by OPA will be compared with the `expect` assertion in the test;

6. all tests' results are then collated in a JSON report (`results.json`) and written out to `out/reports`

Tests will be run in parallel, using a number of workers determined by the available CPU cores (up to 70%) which can be changed using the `-workers` flag (using `1` disables running tests in parallel).

**TODO: the process of templatizing the JSON requests is still TBD**

---

# Notes

<sup>1</sup> See ['users_tests.yaml'](examples/tests/users_tests.yaml) for an example.

<sup>2</sup> See [`jwt-opa`](https://github.com/massenz/jwt-opa).
