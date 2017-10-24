package internal

import (
	"testing"
	"time"
)

func dur(d int) time.Duration {
	return time.Duration(d)
}

func TestCalculateApdexZone(t *testing.T) {
	if z := CalculateApdexZone(dur(10), dur(1)); z != ApdexSatisfying {
		t.Fatal(z)
	}
	if z := CalculateApdexZone(dur(10), dur(10)); z != ApdexSatisfying {
		t.Fatal(z)
	}
	if z := CalculateApdexZone(dur(10), dur(11)); z != ApdexTolerating {
		t.Fatal(z)
	}
	if z := CalculateApdexZone(dur(10), dur(40)); z != ApdexTolerating {
		t.Fatal(z)
	}
	if z := CalculateApdexZone(dur(10), dur(41)); z != ApdexFailing {
		t.Fatal(z)
	}
	if z := CalculateApdexZone(dur(10), dur(100)); z != ApdexFailing {
		t.Fatal(z)
	}
}

func TestApdexLabel(t *testing.T) {
	if out := ApdexSatisfying.label(); "S" != out {
		t.Fatal(out)
	}
	if out := ApdexTolerating.label(); "T" != out {
		t.Fatal(out)
	}
	if out := ApdexFailing.label(); "F" != out {
		t.Fatal(out)
	}
	if out := ApdexNone.label(); "" != out {
		t.Fatal(out)
	}
}
