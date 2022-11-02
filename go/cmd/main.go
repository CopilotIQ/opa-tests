// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22
package main

import (
    "flag"
    . "gentests/gentests"
    "github.com/massenz/slf4go/logging"
    "os"
    "strings"
)

const (
    Manifest = "manifest.json"
    Tests    = "example/tests"
    Dest     = "out/tests"
)

var (
  // Filled in during build
  Release string
)

func main() {
    manifest := flag.String("manifest", Manifest, "Path to the manifest file")
    dest := flag.String("d", Dest, "Path to the destination directory")
    debug := flag.Bool("v", false, "Enable verbose logging")

    flag.Parse()

    // Path to the tests directory
    tests := flag.Arg(0)
    if tests == "" {
        tests = Tests
    }

    log := logging.NewLog("opa-tests")
    if *debug {
        logging.RootLog.Level = logging.DEBUG
        log.Level = logging.DEBUG
    }
    m := ReadManifest(*manifest)
    log.Info("Generating Testcases from: %s -- Bundle rev. %s", tests, m.Revision)
    log.Info("Tests will be generated into: %s", *dest)

    var tg = &TestGenerator{
        SourceDir: tests,
        AllowDir:  strings.Join([]string{*dest, "allow"}, string(os.PathSeparator)),
        DenyDir:   strings.Join([]string{*dest, "deny"}, string(os.PathSeparator)),
    }

    if err := tg.Generate(); err != nil {
        panic(err)
    }
    log.Info("SUCCESS - All tests generated")
}
