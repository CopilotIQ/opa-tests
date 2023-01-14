// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22
package main

import (
	"flag"
	"fmt"
	. "github.com/CopilotIQ/opa-tests/common"
	. "github.com/CopilotIQ/opa-tests/gentests"
	. "github.com/CopilotIQ/opa-tests/workers"
	slf4go "github.com/massenz/slf4go/logging"
)

const (
	Manifest = "manifest.json"
)

var (
	// Release is filled in during build
	Release  string
	ProgName string
)

func main() {
	manifest := flag.String("manifest", Manifest, "Path to the manifest file")
	debug := flag.Bool("v", false, "Enable verbose logging")
	opaUrl := flag.String("opa", "http://localhost:8181", "The URL for the OPA server")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [-v] [-opa URL] [-manifest MANIFEST] TESTS\n\n", ProgName)
		//goland:noinspection GoPrintFunctions
		fmt.Println("TESTS    the directory containing the YAML test cases")
		flag.PrintDefaults()
	}
	flag.Parse()

	log := slf4go.NewLog("opa-tests")
	log.Info("Rego Test Generation Utility - rev. %s", Release)

	// Path to the tests directory
	srcDir := flag.Arg(0)
	if srcDir == "" {
		flag.Usage()
		log.Fatal(fmt.Errorf("missing tests directory"))
	}

	if *debug {
		log.Level = slf4go.DEBUG
	}
	m := ReadManifest(*manifest)
	log.Info("Generating Testcases from: %s -- Bundle rev. %s", srcDir, m.Revision)
	tests, err := Generate(srcDir)
	if err != nil {
		log.Error("cannot read test cases: %s", err)
	}
	log.Info("SUCCESS - All tests generated")

	dataChan := make(chan Request)
	go func() {
		err := SendData(*opaUrl, dataChan)
		if err != nil {
			log.Error("error sending requests to OPA server: %v", err)
		}
	}()
	for _, req := range tests {
		dataChan <- req
	}
	// Once you're done sending data, close the channel
	close(dataChan)
}
