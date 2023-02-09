// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package testing

import "sync"

// A BundleManifest describes the `Bundle` to the OPA server
// We only use it for informational purposes during tests execution.
type BundleManifest struct {
	Revision string            `json:"revision"`
	Roots    []string          `json:"roots"`
	Metadata map[string]string `json:"metadata"`
}

// A Resource describes the access request from the test users (
// the "sub") as the tuple of a Method (the action on the resource) and the Path (
// the actual entity being accessed).
// This is what the Rego policies will evaluate,
// against the user Roles (carried in the Token) to assess whether to allow or deny access (
// and that the Test asserts against its Expectation)
type Resource struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// A Request is what is typically sent from a REST API server that requires
// the user (authenticated by the `Token`) to be authorized to access the `Resource`
type Request struct {
	// A base-64-encoded JWT
	Token string `json:"api_token"`

	// The Resource that the Token's Subject is trying to access
	Resource Resource `json:"resource"`
}

// A TestBody is the JSON that will be sent to OPA to evaluate the Policy.
type TestBody struct {
	Input Request `json:"input"`
}

// A JwtBody describes the contents ("claims") that will be included in the JSON body,
// and that will be part of the OPA Request: the Token (encoded as a JWT) describes the Subject
// ( "sub") who acts as the original sender of the request, who has been
// assigned a number of Roles.
//
// Based on the Policy, one (or more) of the Roles may allow permissions to access the Resource,
// and the Test asserts truth of falsity of this statement.
type JwtBody struct {
	Subject string   `json:"sub" yaml:"sub"`
	Roles   []string `json:"roles" yaml:"roles"`
	Issuer  string   `json:"iss" yaml:"iss"`
}

// Target defines the policy that we want to test with the Testcase
// and will be ultimately used to construct the OPA endpoint to use
// for the Test
type Target struct {
	// Package matches the Rego module `package` keyword
	Package string `yaml:"package"`

	// The Policy matches the Rego module rule that we are testing; there
	// can only be one Policy per Testcase, hence to test different rules
	// in the same Package, you will need to create several Testcase
	Policy string `yaml:"policy"`
}

// A Test is a single assertion made against the Rego policies,
// with an expectation of success or failure (
// depending on whether we are testing an `allow` or `deny` scenario).
//
// Test are grouped in Testcase units and will map one-to-one to OPA server HTTP Request objects,
// invoked against the Target (policy).
type Test struct {
	Name     string   `yaml:"name"`
	Expect   bool     `yaml:"expect"`
	Token    JwtBody  `yaml:"token"`
	Resource Resource `yaml:"resource"`
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
	Iss    string `yaml:"iss"`
	Target Target `yaml:"target"`
	Tests  []Test `yaml:"tests"`
}

// A TestcaseTemplate is the contents of a YAML (
// or a section thereof) and contains the full description of a `Testcase`
// that will be generated.
type TestcaseTemplate struct {
	Body Testcase `yaml:"testcase"`
}

// The TestUnit is the culmination of the test generation,
// and is the unit that is evaluated by each of the workers, run in parallel as goroutines.
// The TestUnit unifies the test subject (the Endpoint),
// the Body of the test (what we are evaluating against the policy defined for the Endpoint) and
// the Expectation (whether this is expected to succeed or fail).
type TestUnit struct {
	Name        string
	Endpoint    string
	Body        TestBody
	Expectation bool
}

// TestReport will collect and report all test results, including failures
// TODO: need a much better reporting
type TestReport struct {
	Succeeded uint
	Failed    uint
	Total     uint

	FailedNames []string

	mutex sync.Mutex
}

func (r *TestReport) IncSuccess() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Total++
	r.Succeeded++
}

func (r *TestReport) ReportFailure(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Total++
	r.Failed++
	r.FailedNames = append(r.FailedNames, name)
}
