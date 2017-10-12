package gostub

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubEnv(t *testing.T) {
	os.Setenv("GOSTUB_T1", "V1")
	os.Setenv("GOSTUB_T2", "V2")
	os.Unsetenv("GOSTUB_NONE")

	stubs := New()

	stubs.SetEnv("GOSTUB_NONE", "a")
	stubs.SetEnv("GOSTUB_T1", "1")
	stubs.SetEnv("GOSTUB_T1", "2")
	stubs.SetEnv("GOSTUB_T1", "3")
	stubs.SetEnv("GOSTUB_T2", "4")
	stubs.UnsetEnv("GOSTUB_T2")

	assert.Equal(t, "3", os.Getenv("GOSTUB_T1"), "Wrong value for T1")
	assert.Equal(t, "", os.Getenv("GOSTUB_T2"), "Wrong value for T2")
	assert.Equal(t, "a", os.Getenv("GOSTUB_NONE"), "Wrong value for NONE")
	stubs.Reset()

	_, ok := os.LookupEnv("GOSTUB_NONE")
	assert.False(t, ok, "NONE should be unset")

	assert.Equal(t, "V1", os.Getenv("GOSTUB_T1"), "Wrong reset value for T1")
	assert.Equal(t, "V2", os.Getenv("GOSTUB_T2"), "Wrong reset value for T2")
}
