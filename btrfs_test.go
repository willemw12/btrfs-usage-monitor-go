package main

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestBtrfsUsage1(t *testing.T) {
	// Get input
	path := "/mnt/btrfs"
	// var freeLimitPercentage uint64 = 60
	freeLimitPercentage := uint64(60)
	cmdOutUsageRaw, err := ioutil.ReadFile("testdata/input-usage-raw1.txt")
	if err != nil {
		log.Fatal(err)
	}
	cmdOutUsageHuman, err := ioutil.ReadFile("testdata/input-usage-human1.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Run test
	usg := usage{}
	usg.extractBtrfsUsageData(cmdOutUsageRaw, cmdOutUsageHuman)
	if err != nil {
		t.Errorf("%v", err)
	}
	warning := usg.getUsageWarning(path, freeLimitPercentage)

	// Check result
	expected, err := ioutil.ReadFile("testdata/output-warning1.txt")
	if err != nil {
		log.Fatal(err)
	}
	if strings.Compare(warning, string(expected)) != 0 {
		t.Errorf("\nExpected: %sGot:      %s", expected, warning)
	}
}
