package sysinfo

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestPhysicalMemoryBytes(t *testing.T) {
	prtconf, err := prtconfMemoryBytes()
	if err != nil {
		t.Fatal(err)
	}

	sysconf, err := PhysicalMemoryBytes()
	if err != nil {
		t.Fatal(err)
	}

	// The pagesize*pages calculation, although standard (the JVM, at least,
	// uses this approach), doesn't match up exactly with the number
	// returned by prtconf.
	if sysconf > prtconf || sysconf < (prtconf-prtconf/20) {
		t.Fatal(prtconf, sysconf)
	}
}

var (
	ptrconfRe = regexp.MustCompile(`[Mm]emory\s*size:\s*([0-9]+)\s*([a-zA-Z]+)`)
)

func prtconfMemoryBytes() (uint64, error) {
	output, err := exec.Command("/usr/sbin/prtconf").Output()
	if err != nil {
		return 0, err
	}

	m := ptrconfRe.FindSubmatch(output)
	if m == nil {
		return 0, errors.New("memory size not found in prtconf output")
	}

	size, err := strconv.ParseUint(string(m[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	switch strings.ToLower(string(m[2])) {
	case "megabytes", "mb":
		return size * 1024 * 1024, nil
	case "kilobytes", "kb":
		return size * 1024, nil
	default:
		return 0, errors.New("couldn't parse memory size in prtconf output")
	}
}
