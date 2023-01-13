// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package common

// A BundleManifest describes the `Bundle` to the OPA server
// We only use it for informational purposes during tests execution.
type BundleManifest struct {
	Revision string            `json:"revision"`
	Roots    []string          `json:"roots"`
	Metadata map[string]string `json:"metadata"`
}

type Resource struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// A Request is what is typically sent from a REST API server that requires
// the user (authenticated by the `Token`) to be authorized to access the `Resource`
type Request struct {
	Token    string   `json:"api_token"`
	Resource Resource `json:"resource"`
}

// A TestBody is the JSON that will be sent to OPA to evaluate the Policy.
type TestBody struct {
	Input Request `json:"input"`
}

type JwtBody struct {
	Subject string   `json:"sub" yaml:"sub"`
	Roles   []string `json:"roles" yaml:"roles"`
	Issuer  string   `json:"iss" yaml:"iss"`
}

// A Testcase is the central part of the application: it describes a coherent
// set of `Tests` that will be evaluated against the deployed `Bundle` of policies
// The `expect` outcome will be validated once the tests are executed, against
// the actual result returned by the OPA server.
type Testcase struct {
	Name string `yaml:"name"`
	Desc string `yaml:"description"`

	// The "iss" claim for the JWT to be generated; can be overridden in a `Test`
	// using the `Token.Issuer` field, if needed.
	Iss   string `yaml:"iss"`
	Tests []Test `yaml:"tests"`
}

type Test struct {
	Name     string   `yaml:"name"`
	Expect   bool     `yaml:"expect"`
	Token    JwtBody  `yaml:"token"`
	Resource Resource `yaml:"resource"`
}

// A TestcaseTemplate is the contents of a YAML (
// or a section thereof) and contains the full description of a `Testcase`
// that will be generated.
type TestcaseTemplate struct {
	Body Testcase `yaml:"testcase"`
}
