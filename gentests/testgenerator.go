// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package gentests

import (
	"github.com/massenz/slf4go/logging"
	"path/filepath"
	"strings"
)

const (
	YamlGlob     = "*.yaml"
	PoliciesGlob = "*.rego"
)

var Log = logging.NewLog("testgen")

func NewRequest(t *Test) Request {
	Log.Debug("Creating Request: %v", *t)
	return Request{
		Token:    NewToken(&t.Token),
		Resource: t.Resource,
	}
}

// Generate all the test cases from the `SourceDir`
func Generate(SourceDir string) ([]TestUnit, error) {
	Log.Debug("Generating test requests from %s", SourceDir)

	// TODO: walk the subtree (instead of just the directory) and modify the test names to
	// 		 reflect the position in the subtree using WalkDir(root string, fn fs.WalkDirFunc)
	files, err := filepath.Glob(filepath.Join(SourceDir, YamlGlob))
	if err != nil {
		return nil, err
	}

	var requests = make([]TestUnit, 0)
	for _, file := range files {
		Log.Debug("- %s", file)
		testcase, err := ReadTestcase(file)
		if err != nil {
			Log.Error("could not read YAML %s: %s", file, err)
			return nil, err
		}
		endpoint := strings.Join([]string{testcase.Target.Package, testcase.Target.Policy}, "/")
		Log.Debug("%s === %s", file, testcase.Name, endpoint)
		for _, test := range testcase.Tests {
			testname := strings.Join([]string{testcase.Name, test.Name}, ".")
			Log.Debug("--- %s", testname)
			Log.Trace("JWT contents: %v", test.Token)
			if test.Token.Issuer == "" {
				test.Token.Issuer = testcase.Iss
			}
			requests = append(requests, TestUnit{
				Name:        testname,
				Endpoint:    endpoint,
				Body:        TestBody{Input: NewRequest(&test)},
				Expectation: test.Expect,
			})
		}
	}
	Log.Info("Generated %d tests", len(requests))
	return requests, nil
}
