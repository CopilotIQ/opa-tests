// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22
package main

import (
	"flag"
	"fmt"
	. "github.com/CopilotIQ/opa-tests/gentests"
	. "github.com/CopilotIQ/opa-tests/gentests/internals"
	slf4go "github.com/massenz/slf4go/logging"
	"math"
	"os"
	"runtime"
	"sync"
)

const (
	Manifest  = "src/main/resources/manifest.json"
	Sources   = "src/main/rego"
	Templates = "src/tests/resources"
	Tests     = "src/tests"
	Out       = "out/reports/results.json"
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
	templates := flag.String("templates", Templates,
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

	Log.Info("generating Bundle rev. %s from %s", m.Revision, *src)
	// TODO: generate Bundle and store where the TestContainer can mount it

	Log.Info("Generating Testcases from: %s", testsDir)
	tests, err := Generate(testsDir)
	if err != nil {
		Log.Error("cannot read test cases: %s", err)
		os.Exit(1)
	}
	if len(tests) == 0 {
		Log.Error("nothing to do")
		os.Exit(1)
	}
	Log.Info("All tests generated")
	Log.Warn("JSON templates in %s support not implemented yet", *templates)

	// TODO: start TestContainer OPA and obtain URL Address
	RunTests(tests, *workers, "http://localhost:8181")

	Log.Info("Test results saved to %s", *out)
	// TODO: collect test results to JSON report
}

func RunTests(tests []TestUnit, workers uint, addr string) {
	dataChan := make(chan TestUnit)
	var wg sync.WaitGroup
	if workers == 0 {
		workers = EstimateWorkers()
	}
	if workers > 1 {
		Log.Info("running %d parallel test runners", workers)
	} else {
		Log.Warn("running single-core, execution will be slower")
	}
	// TODO: create a TestReport to pass in to each coroutine to populate (protect w Mutex)
	for i := uint(0); i < workers; i++ {
		wg.Add(1)
		go func() {
			err := SendData(addr, dataChan)
			if err != nil {
				Log.Error("error sending requests to OPA server: %v", err)
			}
			wg.Done()
		}()
	}
	for _, req := range tests {
		dataChan <- req
	}
	// Once you're done sending data, close the channel
	close(dataChan)
	wg.Wait()
}

func EstimateWorkers() uint {
	cores := runtime.NumCPU()
	if cores > 3 {
		return uint(math.Ceil(0.7 * float64(cores)))
	}
	return 1
}
