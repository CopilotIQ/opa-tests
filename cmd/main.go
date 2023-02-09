// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/CopilotIQ/opa-tests/testing"
	. "github.com/CopilotIQ/opa-tests/testing/internals"
	slf4go "github.com/massenz/slf4go/logging"
	"os"
	"path/filepath"
	"time"
)

const (
	Manifest  = "src/main/resources/manifest.json"
	Sources   = "src/main/rego"
	Templates = "src/tests/resources"
	Tests     = "src/tests"
	Out       = "out/reports/results.json"

	DefaultOpaContainerStartTimeout = 1 * time.Minute
)

var (
	// Release is filled in during build
	Release  string
	ProgName string
)

func main() {

	manifest := flag.String("manifest", Manifest, "Path to the manifest file")
	src := flag.String("src", Sources, "Path to policies (Rego)")
	out := flag.String("out", Out, "Path to test results report")
	workers := flag.Uint("workers", 0, "Number of parallel threads to run")
	// TODO: replace default with `Templates` once JSON templates are implemented
	templates := flag.String("templates", "",
		"Directory containing (optional) Golang templates for the test requests' JSON body")
	debug := flag.Bool("v", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [-v] [-manifest MANIFEST] [-src SRC] "+
			"[-templates TEMPLATES] [-out REPORT] [TESTS]\n\n", ProgName)
		flag.PrintDefaults()
		//goland:noinspection GoPrintFunctions
		fmt.Printf("\nRuns all the tests in the TESTS folder (default \"%s\")\n", Tests)
	}
	flag.Parse()

	if *debug {
		Log.Level = slf4go.DEBUG
	}
	Log.Info("Rego Test Generation Utility - rev. %s", Release)

	// Path to the tests directory
	testsDir := flag.Arg(0)
	if testsDir == "" {
		testsDir = Tests
	}

	m := ReadManifest(*manifest)

	Log.Info("Generating Bundle rev. %s from %s", m.Revision, *src)
	bundle, err := CreateBundle(*manifest, *src)
	if err != nil {
		Log.Fatal(err)
	}
	defer os.Remove(bundle)
	Log.Debug("bundle %s created", bundle)

	Log.Info("Generating Testcases from: %s", testsDir)
	tests, err := Generate(testsDir)
	if err != nil {
		Log.Fatal(fmt.Errorf("cannot read test cases: %s", err))
	}
	if len(tests) == 0 {
		Log.Fatal(fmt.Errorf("nothing to do"))
	}
	Log.Info("All tests generated")

	if *templates != "" {
		Log.Fatal(fmt.Errorf("JSON templates support not implemented yet"))
	}

	EnsureReportDir(*out)
	file, _ := os.Create(*out)
	defer file.Close()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOpaContainerStartTimeout)
	defer cancel()
	server, err := NewOpaContainer(ctx, bundle)
	if err != nil {
		Log.Fatal(err)
	}
	name, _ := server.Container.Name(ctx)
	Log.Info("OPA Server started (container: %s)", name)

	report := RunTests(tests, *workers, server.Address)
	err = server.Container.Terminate(ctx)
	if err != nil {
		Log.Error("failed to stop OPA container: %v", err)
	}
	encoder := json.NewEncoder(file)
	err = encoder.Encode(report)
	if err != nil {
		Log.Fatal(err)
	}
	elapsed := time.Since(start)
	b, _ := json.MarshalIndent(report, "", "    ")
	fmt.Println(string(b))
	Log.Info("Took %v -- Test results saved to %s", elapsed, *out)
}

func EnsureReportDir(report string) {
	dir, _ := filepath.Split(report)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			Log.Debug("creating test results directory %s", dir)
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				Log.Debug("failed to create directory %s: %v", dir, err)
				Log.Fatal(err)
			}
		}
	}
}
