// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package testing

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"os"
)

func ReadManifest(path string) *BundleManifest {
	jsonManifest, err := os.Open(path)
	if err != nil {
		Log.Fatal(err)
	}
	var manifest BundleManifest
	err = json.NewDecoder(jsonManifest).Decode(&manifest)
	if err != nil {
		Log.Error("cannot decode Manifest %s: %s", path, err)
		return nil
	}
	return &manifest
}

func ReadTestcase(path string) (*Testcase, error) {
	yamlTestcase, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var template TestcaseTemplate
	if err := yaml.NewDecoder(yamlTestcase).Decode(&template); err != nil {
		Log.Error("cannot decode Testcase %s: %s", path, err)
		return nil, err
	}
	return &template.Body, nil
}
