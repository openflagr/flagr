package sysinfo

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestDockerIDCrossAgent(t *testing.T) {
	var testCases []struct {
		File string `json:"filename"`
		ID   string `json:"containerId"`
	}

	dir := "docker_container_id"
	err := crossagent.ReadJSON(filepath.Join(dir, "cases.json"), &testCases)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range testCases {
		file := filepath.Join(dir, test.File)
		input, err := crossagent.ReadFile(file)
		if err != nil {
			t.Error(err)
			continue
		}

		got, _ := parseDockerID(bytes.NewReader(input))
		if got != test.ID {
			t.Errorf("%s != %s", got, test.ID)
		}
	}
}

func TestDockerIDValidation(t *testing.T) {
	err := validateDockerID("baaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1239")
	if nil != err {
		t.Error("Validation should pass with a 64-character hex string.")
	}
	err = validateDockerID("39ffbba")
	if nil == err {
		t.Error("Validation should have failed with short string.")
	}
	err = validateDockerID("z000000000000000000000000000000000000000000000000100000000000000")
	if nil == err {
		t.Error("Validation should have failed with non-hex characters.")
	}
}
