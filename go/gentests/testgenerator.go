// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package gentests

import (
	. "github.com/CopilotIQ/opa-tests/common"
	"github.com/massenz/slf4go/logging"
	"path/filepath"
)

const (
	YamlGlob = "*.yaml"
)

var log = logging.RootLog

func NewRequest(t *Test) Request {
	log.Debug("Creating Request: %v", *t)
	return Request{
		Token:    NewToken(&t.Token),
		Resource: t.Resource,
	}
}

// Generate all the test cases from the `SourceDir`
func Generate(SourceDir string) ([]Request, error) {
	log.Debug("Generating test requests from %s", SourceDir)

	// TODO: walk the subtree (instead of just the directory) and modify the test names to
	// reflect the position in the subtree.
	//     filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
	//        if !info.IsDir() {
	//            // generate TestCase
	//        }
	//        return nil
	//    })
	files, err := filepath.Glob(filepath.Join(SourceDir, YamlGlob))
	if err != nil {
		return nil, err
	}

	var requests = make([]Request, len(files))
	for _, file := range files {
		log.Debug("- %s", file)
		testcase, err := ReadTestcase(file)
		if err != nil {
			log.Error("could not read YAML %s: %s", file, err)
			return nil, err
		}
		log.Debug("Creating Test (%s): %s", file, testcase.Name)
		for _, test := range testcase.Tests {
			log.Debug("JWT contents: %v", test.Token)
			if test.Token.Issuer == "" {
				test.Token.Issuer = testcase.Iss
			}
			requests = append(requests, NewRequest(&test))
		}
	}
	return requests, nil
}
