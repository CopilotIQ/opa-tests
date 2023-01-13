// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package gentests

import (
	"encoding/json"
	"github.com/CopilotIQ/opa-tests/common"
	"gopkg.in/yaml.v2"
	"os"
)

func ReadManifest(path string) *common.BundleManifest {
	jsonManifest, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	var manifest common.BundleManifest
	err = json.NewDecoder(jsonManifest).Decode(&manifest)
	if err != nil {
		log.Error("cannot decode Manifest %s: %s", path, err)
		return nil
	}
	return &manifest
}

func ReadTestcase(path string) (*common.Testcase, error) {
	yamlTestcase, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var template common.TestcaseTemplate
	if err := yaml.NewDecoder(yamlTestcase).Decode(&template); err != nil {
		log.Error("cannot decode Testcase %s: %s", path, err)
		return nil, err
	}
	return &template.Body, nil
}
