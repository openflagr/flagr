package internal

import (
	"net/url"
	"strings"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestSafeURLNil(t *testing.T) {
	if out := SafeURL(nil); "" != out {
		t.Error(out)
	}
}

func TestSafeURL(t *testing.T) {
	var testcases []struct {
		Testname string `json:"testname"`
		Expect   string `json:"expected"`
		Input    string `json:"input"`
	}

	err := crossagent.ReadJSON("url_clean.json", &testcases)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		if strings.Contains(tc.Input, ";") {
			// This test case was over defensive:
			// http://www.ietf.org/rfc/rfc3986.txt
			continue
		}

		// Only use testcases which have a scheme, otherwise the urls
		// may not be valid and may not be correctly handled by
		// url.Parse.
		if strings.HasPrefix(tc.Input, "p:") {
			u, err := url.Parse(tc.Input)
			if nil != err {
				t.Error(tc.Testname, tc.Input, err)
				continue
			}
			out := SafeURL(u)
			if out != tc.Expect {
				t.Error(tc.Testname, tc.Input, tc.Expect)
			}
		}
	}
}

func TestSafeURLFromString(t *testing.T) {
	out := SafeURLFromString(`http://localhost:8000/hello?zip=zap`)
	if `http://localhost:8000/hello` != out {
		t.Error(out)
	}
	out = SafeURLFromString("?????")
	if "" != out {
		t.Error(out)
	}
}

func TestHostFromURL(t *testing.T) {
	u, err := url.Parse("http://example.com/zip/zap?secret=shh")
	if nil != err {
		t.Fatal(err)
	}
	host := HostFromURL(u)
	if host != "example.com" {
		t.Error(host)
	}
	host = HostFromURL(nil)
	if host != "" {
		t.Error(host)
	}
	u, err = url.Parse("scheme:opaque")
	if nil != err {
		t.Fatal(err)
	}
	host = HostFromURL(u)
	if host != "opaque" {
		t.Error(host)
	}
}
