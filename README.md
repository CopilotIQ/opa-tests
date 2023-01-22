# Open Policy Agent (OPA) Integration Tests Framework


### Copyright & Licensing

**The code is copyright (c) 2023 CopilotIQ Inc. All rights reserved**

*This code is licensed under the terms of the Apache License 2.0, see LICENSE for more details*

**NOTE THIS WILL BECOME A PUBLIC REPOSITORY AND NO PROPRIETARY CopilotIQ CODE SHOULD BE STORED HERE**


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

A `Testcase` (a self-contained YAML descriptor<sup>2</sup>) describes a set of `Test`s which will be run in sequence; each `Test` will result in a JSON payload which, depending on the value of `expect` will be generated in either the `allow` or `deny` subfolder of `src/test/resouces` and will respectively expect a `{"result": true | false}` when the tests are run, from OPA.

Each OPA request carries along a JWT<sup>3</sup>, which will be encoded and signed by the generator, and sent along with the request.

Each JWT carries a `user` and an array or `roles`<sup>4</sup> which will be used to authorize the request.

A `policy request` essentially asks OPA to assert whether the given `role` is allowed to perform the given "action" (defined by the HTTP `method`) on the "entity" (encoded in the `path`).

In this respect a policy `P` defines a `mapping` from a set `{role, action, entity}` to `boolean`:

    P: {role, action, entity} -> boolean

Each `Test` in a `Testcase` then defines an invariant (or, if you will, an assertion) upon the whole set of policies `{P}` as it regards one specific mapping.

### Reference

`TODO: detailed explanation of each field in the testcase YAML`


# Execute Tests

In its simplest form, tests from a `path/to/tests` directory can be run against a running OPA server (see [`run-opa`](run-opa) for an example of how to do it) which has loaded the Rego policies that we wish to test from a "bundle" (see [Bundle](#bundle) for an example of how to generate one).

`opa-test -manifest policies.json -opa http://localhost:8089 path/to/tests`

will load Manifest metadata from the `some/path/policies.json` file (by default uses `manifest.json` in the current directory) and then generate the Testcases (`*.yaml`) from the `path/to/tests`directory.

If you want to know how stuff works, keep reading.

## Testing details

Locally (and for testing) we use a simple configuration which loads the bundle directly from the filesystem, unlike the one deployed in AWS's EKS<sup>5</sup>, so that we do not need to upload/download policies from S3 bucket every time there is a change; however this means that the container needs restarting every time you change one of the Rego policies and want the new version to be picked up<sup>6</sup>.

Essentially, the cycle is "edit policies / bundle / restart OPA / test" using the respective scripts<sup>7</sup>:

```
# To build the latest policy bundle (to the default out/bundles directory)
./bundle path/to/policies

# To run (or restart) the OPA Server locally; the version
# will be emitted by the bundle script
./run-opa out/bundles/authz-{version}.tar.gz

# To test the policies (we recommend adding opa-test to your PATH)
opa-test -manifest src/rego/resources/manifest.json \
    -opa http://localhost:8181 \
    src/rego/tests
```

**TO BE IMPLEMENTED**

A lot of the above will be simplified, by encapsulating the whole functionality in the `opa-test` binary:

```
opa-test -manifest src/rego/resources/manifest.json \
    -policies src/rego/main -out report.json -p \
    src/rego/tests
```

The Go binary will combine all the functionality of building the policies bundle, starting the OPA container, generating the tests and emitting the report.

Optionally, the `-p` flag will tell the runner to run tests in parallel to speed up execution.

Finally, assuming the default structure of the repository to be gradle-like, all the above would be default values, so the same as above could be achieved simply with:

```
opa-test -p
```

and the `report.json` generated in the default `out/reports`.



# Optional Components

## Bundle

OPA supports the concept of `bundle` and the server will mount the bundle from the filesystem, as generated from the Rego policies we are testing; while the format of the Bundle (and how it is generated) is largely irrelevant for `opa-test`, and the user is free (in fact, encouraged) to use their own automation, nonetheless we also provide a utility script that will build an OPA-compliant bundle from a `manifest` and a set of Rego files.


```
./bundle -h
```

to see usage instructions.

```
└─( ./bundle examples/policies out/bundles examples/policies/manifest.json
Bundle out/bundles/authz-0.6.26.tar.gz exists. Overwrite [y/n]? y
Created authz-0.6.26.tar.gz (Rev. 0.6.26) in out/bundles
```

The command above will package the Rego files in a tar gzipped file in `out/bundles`, named something like `authz-0.6.26.tar.gz` (the bundle version is derived from [the Manifest](examples/policies/manifest.json))

## Common Utilities

The `bundle` script uses the [Common Utilities](https://github.com/massenz/common-utils) helpers (in particular the `utils` functions and `parse-args`): please see that repository for instructions on how to [install a recent release](https://github.com/massenz/common-utils#usage) and configure your environment `$UTILS_DIR` variable to point to that folder.

---

# Notes

<sup>1</sup> This is however taken care of in the `test` script, and is not necessary.

<sup>2</sup> See ['events_tests.yaml'](src/test/tests/events_tests.yaml) for an example.

<sup>3</sup> See [`jwt-opa`](https://github.com/massenz/jwt-opa).

<sup>6</sup> See [here](run#L34)

<sup>7</sup> The bundle does not get reloaded if the underlying file changes; upon running `bundle` you will need to restart the container using `run-opa`.

<sup>8</sup> Note the use of `/com.example` instead of the full `/com.example/allow`
