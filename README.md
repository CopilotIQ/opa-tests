# Open Policy Agent (OPA) Authorization Policies


### Copyright & Licensing

**The code is copyright (c) 2021-2022 CopilotIQ Inc. All rights reserved**

**This code is licensed under the terms of the Apache License 2.0, see LICENSE for more details**

# Motivation

This repository contains a policy validation framework for [Open Policy Agent](https://opa.io) server policies.

**NOTE THIS WILL BECOME A PUBLIC REPOSITORY AND NO PROPRIETARY CopilotIQ CODE SHOULD BE STORED HERE**

`TODO: this README is work in progress`

## Build the Bundle

Our OPA server will pull the "bundle" from an S3 bucket specified in the [configuration](docker/opa-config.yaml) file; we must build the bundle from the Rego policies (in `src/main/rego`) and then upload it to S3:

```
./bundle
```

The command above will package the Rego files in a tar gzipped file in `out/bundles`, named something like `authz-0.6.2.tar.gz` (the bundle version is derived from [the Manifest](src/main/resources/manifest.json))

## Test Policies

This repository `run` and `test` scripts use the [Common Utilities](https://github.com/massenz/common-utils) helpers (in particular the `utils` functions): please see that repository for instructions on how to [install a recent release](https://github.com/massenz/common-utils#usage) and configure your environment `$UTILS_DIR` variable to point to that folder.

## Build Tests

All Policies Tests are generated following the `Testcase` definitions (in YAML) in the `src/test/tests` folder, and running the `out/bin/gentests` test generator.

The Generator itself is written in Go and built using the following<sup>1</sup> commands:

```shell
gentests=out/bin/gentests
pushd src/main/go
go build -o ../../../${gentests} cmd/generate-tests.go
popd
```

#### Testcases

A `Testcase` (a self-contained YAML descriptor<sup>2</sup>) describes a set of `Test`s which will be run in sequence; each `Test` will result in a JSON payload which, depending on the value of `expect` will be generated in either the `allow` or `deny` subfolder of `src/test/resouces` and will respectively expect a `{"result": true | false}` when the tests are run, from OPA.

Each OPA request carries along a JWT<sup>3</sup>, which will be encoded and signed by the generator, and sent along with the request.

Each JWT carries a `user` and an array or `roles`<sup>4</sup> which will be used to authorize the request.

A `policy request` essentially asks OPA to assert whether the given `role` is allowed to perform the given "action" (defined by the HTTP `method`) on the "entity" (encoded in the `path`).

In this respect a policy `P` defines a `mapping` from a set `{role, action, entity}` to `boolean`:

    P: {role, action, entity} -> boolean

Each `Test` in a `Testcase` then defines an invariant (or, if you will, an assertion) upon the whole set of policies `{P}` as it regards one specific mapping.

## Run Tests

TL;DR: to run tests use `./test` - that's it.
Additionally, if you have just modified the `Rego` policies and rebuilt them with `./bundle`, use the `-f` flag to force the local container restart.

Bottom line, use:

    ./bundle && ./test -f

and that's pretty much all there is to it.

If you want to know how stuff works, keep reading.

#### Testing details

Locally (and for testing) we use a simple configuration which loads the bundle directly from the filesystem, unlike the one deployed in AWS's EKS<sup>5</sup>, so that we do not need to upload/download policies from S3 bucket every time there is a change; however this means that the container needs restarting every time you change one of the Rego policies and want the new version to be picked up:<sup>6</sup>

```shell
docker run --rm -d -p ${port}:8181 --name opa \
  -v ${policies}:${bundle} \
  openpolicyagent/opa:${version} run --server --addr :8181 ${bundle}
```
*Notice the `-v` flag*

Essentially, the cycle is "bundle / run / test" using the respective scripts<sup>7</sup>:

```shell
# To build the latest policy bundle, using
# the `revision` from the manifest.json
./bundle

# To run (or restart) the OPA Server locally:
./run

# To test the policies against payloads in
# the src/test/resources folders
./test
```

Much more succinctly simply use:

    ./bundle && ./test

Each of the above scripts tries to act sensibly (for example, `test` will run the server if none is running, and `run` will restart an already running one).


### View full Policy output

To debug issue, it is usually helpful to see the full output of the policies' evaluation; this can be done with:<sup>8</sup>

```shell
export opa=:8181
http ${opa}/v1/data/copilotiq @src/test/resources/allow/admin-get-users.json
```

---

# Notes

<sup>1</sup> This is however taken care of in the `test` script, and is not necessary.

<sup>2</sup> See ['events_tests.yaml'](src/test/tests/events_tests.yaml) for an example.

<sup>3</sup> See [`jwt-opa`](https://github.com/massenz/jwt-opa).

<sup>6</sup> See [here](run#L34)

<sup>7</sup> The bundle does not get reloaded if the underlying file changes; upon running `bundle` you will need to restart the container using `run`.

<sup>8</sup> Note the use of `/com.example` instead of the full `/com.example/allow`
