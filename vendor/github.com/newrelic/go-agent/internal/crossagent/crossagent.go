package crossagent

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var (
	crossAgentDir = func() string {
		if s := os.Getenv("NEW_RELIC_CROSS_AGENT_TESTS"); s != "" {
			return s
		}
		_, here, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(here), "cross_agent_tests")
	}()
)

// ReadFile reads a file from the crossagent tests directory given as with
// ioutil.ReadFile.
func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(crossAgentDir, name))
}

// ReadJSON takes the name of a file and parses it using JSON.Unmarshal into
// the interface given.
func ReadJSON(name string, v interface{}) error {
	data, err := ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ReadDir reads a directory relative to crossagent tests and returns an array
// of absolute filepaths of the files in that directory.
func ReadDir(name string) ([]string, error) {
	dir := filepath.Join(crossAgentDir, name)

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, info := range entries {
		if !info.IsDir() {
			files = append(files, filepath.Join(dir, info.Name()))
		}
	}
	return files, nil
}
