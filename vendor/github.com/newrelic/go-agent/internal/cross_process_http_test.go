package internal

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/newrelic/go-agent/internal/cat"
)

func TestTxnCrossProcessInitFromHTTPRequest(t *testing.T) {
	txp := &TxnCrossProcess{}
	if err := txp.InitFromHTTPRequest(true, replyAccountOne, nil); err != nil {
		t.Errorf("got error while initialising without a request: %v", err)
	}
	if txp.IsInbound() {
		t.Error("inbound CAT enabled even though there was no request")
	}

	txp = &TxnCrossProcess{}
	req, err := http.NewRequest("GET", "http://foo.bar/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := txp.InitFromHTTPRequest(true, replyAccountOne, req); err != nil {
		t.Errorf("got error while initialising with an empty request: %v", err)
	}
	if txp.IsInbound() {
		t.Error("inbound CAT enabled even though there was no metadata in the request")
	}

	txp = &TxnCrossProcess{}
	req, err = http.NewRequest("GET", "http://foo.bar/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(cat.NewRelicIDName, mustObfuscate(`1#1`, "foo"))
	req.Header.Add(cat.NewRelicTxnName, mustObfuscate(`["abcdefgh",false,"12345678","b95be233"]`, "foo"))
	if err := txp.InitFromHTTPRequest(true, replyAccountOne, req); err != nil {
		t.Errorf("got error while initialising with an inbound CAT request: %v", err)
	}
	if !txp.IsInbound() {
		t.Error("inbound CAT disabled even though there was metadata in the request")
	}
	if txp.ClientID != "1#1" {
		t.Errorf("incorrect ClientID: %s", txp.ClientID)
	}
	if txp.ReferringTxnGUID != "abcdefgh" {
		t.Errorf("incorrect ReferringTxnGUID: %s", txp.ReferringTxnGUID)
	}
	if txp.TripID != "12345678" {
		t.Errorf("incorrect TripID: %s", txp.TripID)
	}
	if txp.ReferringPathHash != "b95be233" {
		t.Errorf("incorrect ReferringPathHash: %s", txp.ReferringPathHash)
	}
}

func TestAppDataToHTTPHeader(t *testing.T) {
	header := AppDataToHTTPHeader("")
	if len(header) != 0 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}

	header = AppDataToHTTPHeader("foo")
	if len(header) != 1 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}
	if actual := header.Get(cat.NewRelicAppDataName); actual != "foo" {
		t.Errorf("unexpected header value: %s", actual)
	}
}

func TestHTTPHeaderToAppData(t *testing.T) {
	if appData := HTTPHeaderToAppData(nil); appData != "" {
		t.Errorf("unexpected app data: %s", appData)
	}

	header := http.Header{}
	if appData := HTTPHeaderToAppData(header); appData != "" {
		t.Errorf("unexpected app data: %s", appData)
	}

	header.Add("X-Foo", "bar")
	if appData := HTTPHeaderToAppData(header); appData != "" {
		t.Errorf("unexpected app data: %s", appData)
	}

	header.Add(cat.NewRelicAppDataName, "foo")
	if appData := HTTPHeaderToAppData(header); appData != "foo" {
		t.Errorf("unexpected app data: %s", appData)
	}
}

func TestHTTPHeaderToMetadata(t *testing.T) {
	if metadata := HTTPHeaderToMetadata(nil); !reflect.DeepEqual(metadata, CrossProcessMetadata{}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}

	header := http.Header{}
	if metadata := HTTPHeaderToMetadata(header); !reflect.DeepEqual(metadata, CrossProcessMetadata{}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}

	header.Add("X-Foo", "bar")
	if metadata := HTTPHeaderToMetadata(header); !reflect.DeepEqual(metadata, CrossProcessMetadata{}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}

	header.Add(cat.NewRelicIDName, "id")
	if metadata := HTTPHeaderToMetadata(header); !reflect.DeepEqual(metadata, CrossProcessMetadata{
		ID: "id",
	}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}

	header.Add(cat.NewRelicTxnName, "txn")
	if metadata := HTTPHeaderToMetadata(header); !reflect.DeepEqual(metadata, CrossProcessMetadata{
		ID:      "id",
		TxnData: "txn",
	}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}

	header.Add(cat.NewRelicSyntheticsName, "synth")
	if metadata := HTTPHeaderToMetadata(header); !reflect.DeepEqual(metadata, CrossProcessMetadata{
		ID:         "id",
		TxnData:    "txn",
		Synthetics: "synth",
	}) {
		t.Errorf("unexpected metadata: %v", metadata)
	}
}

func TestMetadataToHTTPHeader(t *testing.T) {
	metadata := CrossProcessMetadata{}

	header := MetadataToHTTPHeader(metadata)
	if len(header) != 0 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}

	metadata.ID = "id"
	header = MetadataToHTTPHeader(metadata)
	if len(header) != 1 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}
	if actual := header.Get(cat.NewRelicIDName); actual != "id" {
		t.Errorf("unexpected header value: %s", actual)
	}

	metadata.TxnData = "txn"
	header = MetadataToHTTPHeader(metadata)
	if len(header) != 2 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}
	if actual := header.Get(cat.NewRelicIDName); actual != "id" {
		t.Errorf("unexpected header value: %s", actual)
	}
	if actual := header.Get(cat.NewRelicTxnName); actual != "txn" {
		t.Errorf("unexpected header value: %s", actual)
	}

	metadata.Synthetics = "synth"
	header = MetadataToHTTPHeader(metadata)
	if len(header) != 3 {
		t.Errorf("unexpected number of header elements: %d", len(header))
	}
	if actual := header.Get(cat.NewRelicIDName); actual != "id" {
		t.Errorf("unexpected header value: %s", actual)
	}
	if actual := header.Get(cat.NewRelicTxnName); actual != "txn" {
		t.Errorf("unexpected header value: %s", actual)
	}
	if actual := header.Get(cat.NewRelicSyntheticsName); actual != "synth" {
		t.Errorf("unexpected header value: %s", actual)
	}
}
