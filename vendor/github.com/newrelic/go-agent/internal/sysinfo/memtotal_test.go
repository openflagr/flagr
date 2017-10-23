package sysinfo

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestMemTotal(t *testing.T) {
	var fileRe = regexp.MustCompile(`meminfo_([0-9]+)MB.txt$`)
	var ignoreFile = regexp.MustCompile(`README\.md$`)

	testCases, err := crossagent.ReadDir("proc_meminfo")
	if err != nil {
		t.Fatal(err)
	}

	for _, testFile := range testCases {
		if ignoreFile.MatchString(testFile) {
			continue
		}

		matches := fileRe.FindStringSubmatch(testFile)

		if matches == nil || len(matches) < 2 {
			t.Error(testFile, matches)
			continue
		}

		expect, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			t.Error(err)
			continue
		}

		input, err := os.Open(testFile)
		if err != nil {
			t.Error(err)
			continue
		}
		bts, err := parseProcMeminfo(input)
		input.Close()
		mib := BytesToMebibytes(bts)
		if err != nil {
			t.Error(err)
		} else if mib != expect {
			t.Error(bts, expect)
		}
	}
}
