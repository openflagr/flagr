package internal

import (
	"net/http"
	"testing"
	"time"
)

func TestParseQueueTime(t *testing.T) {
	badInput := []string{
		"",
		"nope",
		"t",
		"0",
		"0.0",
		"9999999999999999999999999999999999999999999999999",
		"-1368811467146000",
		"3000000000",
		"3000000000000",
		"900000000",
		"900000000000",
	}
	for _, s := range badInput {
		if qt := parseQueueTime(s); !qt.IsZero() {
			t.Error(s, qt)
		}
	}

	testcases := []struct {
		input  string
		expect int64
	}{
		// Microseconds
		{"1368811467146000", 1368811467},
		// Milliseconds
		{"1368811467146.000", 1368811467},
		{"1368811467146", 1368811467},
		// Seconds
		{"1368811467.146000", 1368811467},
		{"1368811467.146", 1368811467},
		{"1368811467", 1368811467},
	}
	for _, tc := range testcases {
		qt := parseQueueTime(tc.input)
		if qt.Unix() != tc.expect {
			t.Error(tc.input, tc.expect, qt, qt.UnixNano())
		}
	}
}

func TestQueueDuration(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("X-Queue-Start", "1465798814")
	qd := QueueDuration(hdr, time.Unix(1465798816, 0))
	if qd != 2*time.Second {
		t.Error(qd)
	}

	hdr = make(http.Header)
	hdr.Set("X-Request-Start", "1465798814")
	qd = QueueDuration(hdr, time.Unix(1465798816, 0))
	if qd != 2*time.Second {
		t.Error(qd)
	}

	hdr = make(http.Header)
	qd = QueueDuration(hdr, time.Unix(1465798816, 0))
	if qd != 0 {
		t.Error(qd)
	}

	hdr = make(http.Header)
	hdr.Set("X-Request-Start", "invalid-time")
	qd = QueueDuration(hdr, time.Unix(1465798816, 0))
	if qd != 0 {
		t.Error(qd)
	}

	hdr = make(http.Header)
	hdr.Set("X-Queue-Start", "t=1465798814")
	qd = QueueDuration(hdr, time.Unix(1465798816, 0))
	if qd != 2*time.Second {
		t.Error(qd)
	}

	// incorrect time order
	hdr = make(http.Header)
	hdr.Set("X-Queue-Start", "t=1465798816")
	qd = QueueDuration(hdr, time.Unix(1465798814, 0))
	if qd != 0 {
		t.Error(qd)
	}
}
