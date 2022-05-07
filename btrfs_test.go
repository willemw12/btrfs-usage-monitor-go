package main

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

const defaultPath = "/mnt/btrfs"

type TestData struct {
	path                string
	rawInputFile        string
	humanInputFile      string
	freeLimitPercentage uint64

	expectedOutputFile string
}

func TestBtrfsUsage(t *testing.T) {
	var (
		dataSet = []TestData{
			{defaultPath, "testdata/input-usage-raw-zero-size.txt", "testdata/input-usage-human-zero-size.txt", 0, "testdata/output-warning-zero-size.txt"},
			{defaultPath, "testdata/input-usage-raw1.txt", "testdata/input-usage-human1.txt", 60, "testdata/output-warning1.txt"},
		}
	)

	for i := 0; i < len(dataSet); i++ {
		data := dataSet[i]

		// Get input.
		freeLimitPercentage := data.freeLimitPercentage
		cmdOutUsageRaw, err := ioutil.ReadFile(data.rawInputFile)
		if err != nil {
			log.Fatal(err)
		}
		cmdOutUsageHuman, err := ioutil.ReadFile(data.humanInputFile)
		if err != nil {
			log.Fatal(err)
		}

		// Run test.
		usg := usage{}
		usg.extractUsage(cmdOutUsageRaw, cmdOutUsageHuman)
		if err != nil {
			t.Fatal(err)
		}
		warning := usg.usageWarning(data.path, freeLimitPercentage)

		// Check result.
		expected, err := ioutil.ReadFile(data.expectedOutputFile)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Compare(warning, string(expected)) != 0 {
			t.Errorf("\nExpected: %sGot:      %s", expected, warning)
		}
	}
}

// func TestBtrfsPanic(t *testing.T) {
// 	var (
// 		dataSet = []TestData{
// 			{defaultPath, "testdata/input-usage-raw-zero-size.txt", "testdata/input-usage-human-zero-size.txt", 0, ""},
// 		}
// 	)
//
// 	defer func() {
// 		if r := recover(); r == nil {
// 			t.Errorf("Expected panic")
// 		}
// 	}()
//
// 	for i := 0; i < len(dataSet); i++ {
// 		data := dataSet[i]
//
// 		// Get input.
// 		freeLimitPercentage := data.freeLimitPercentage
// 		cmdOutUsageRaw, err := ioutil.ReadFile(data.rawInputFile)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		cmdOutUsageHuman, err := ioutil.ReadFile(data.humanInputFile)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
//
// 		// Run test.
// 		usg := usage{}
// 		err = usg.extractUsage(cmdOutUsageRaw, cmdOutUsageHuman)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		usg.usageWarning(data.path, freeLimitPercentage)
// 	}
// }
