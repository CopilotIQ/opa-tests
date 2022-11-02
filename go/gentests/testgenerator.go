// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package gentests

import (
    "encoding/json"
    "github.com/massenz/slf4go/logging"
    "io/ioutil"
    "os"
    "strings"
)

const DefaultIssuer = "copilotiq.com"

type TestGenerator struct {
    SourceDir string
    AllowDir  string
    DenyDir   string

    testcases []*Testcase
}

var log = logging.RootLog

func (tg *Testcase) JsonDestName(test Test) string {
    return strings.Join([]string{tg.Name, test.Name, "json"}, ".")
}

func NewRequest(t *Test) Request {
    log.Debug("Creating Request: %v", *t)
    return Request{
        Token:    NewToken(&t.Token),
        Resource: t.Resource,
    }
}

// Write emits the Test to the given `dest` file
func (t *Test) Write(dest string) error {
    jsonFile, err := os.Create(dest)
    defer jsonFile.Close()
    if err != nil {
        return err
    }
    if err = json.NewEncoder(jsonFile).Encode(TestBody{Input: NewRequest(t)}); err != nil {
        return err
    }
    return nil
}

// Generate all the test cases from the `SourceDir
func (tg *TestGenerator) Generate() error {
    log.Debug("Generating test requests from %s", tg.SourceDir)
    files, err := ioutil.ReadDir(tg.SourceDir)
    if err != nil {
        return err
    }
    log.Debug("Generating JSON payloads in %s and %s", tg.AllowDir, tg.DenyDir)
    for _, tc := range files {
        log.Debug("- %s", tc.Name())
        if !tc.IsDir() && strings.HasSuffix(tc.Name(), ".yaml") {
            testcase, err := readTestcase(
                strings.Join([]string{tg.SourceDir, tc.Name()}, string(os.PathSeparator)))
            if err != nil {
                log.Error("could not read YAML %s: %s", tc.Name(), err)
                return err
            }
            log.Debug("(%s): %s", tc.Name(), testcase.Name)
            for _, test := range testcase.Tests {
                var dest string
                if test.Expect {
                    dest = tg.AllowDir
                } else {
                    dest = tg.DenyDir
                }
                fullpath := strings.Join([]string{dest, testcase.JsonDestName(test)},
                    string(os.PathSeparator))
                log.Debug("Creating Test: %s", fullpath)
                log.Debug("JWT contents: %v", test.Token)
                if test.Token.Issuer == "" {
                    test.Token.Issuer = DefaultIssuer
                }
                err := test.Write(fullpath)
                if err != nil {
                    log.Error("could not write %s: %s", fullpath, err)
                    return err
                }
            }
        }
    }
    return nil
}
