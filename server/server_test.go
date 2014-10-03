package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestInstantaneousDemandRequest(t *testing.T) {
	const body = `
    <?xml version="1.0"?>
    <rainforest macId="0xf0ad4e00ce69" timestamp="1355292588s">
    <InstantaneousDemand>
    <DeviceMacId>0x00158d0000000004</DeviceMacId>
    <MeterMacId>0x00178d0000000004</MeterMacId>
    <TimeStamp>0x185adc1d</TimeStamp>
    <Demand>0x001738</Demand>
    <Multiplier>0x00000001</Multiplier>
    <Divisor>0x000003e8</Divisor>
    <DigitsRight>0x03</DigitsRight>
    <DigitsLeft>0x00</DigitsLeft>
    <SuppressLeadingZero>Y</SuppressLeadingZero>
    </InstantaneousDemand>
    </rainforest>
  `
	record := httptest.NewRecorder()
	req := &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/metrics"},
		Body:   ioutil.NopCloser(strings.NewReader(body)),
	}
	MetricsHandler(record, req)
	if record.Code != 200 {
		t.Errorf("Response got %d not 200", record.Code)
	}
}

func TestGetMetrics(t *testing.T) {
	record := httptest.NewRecorder()
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/metrics"},
	}
	MetricsHandler(record, req)
	if record.Code != 200 {
		t.Errorf("Response was %d not 200", record.Code)
	}
	body, err := ioutil.ReadAll(record.Body)

	matched, err := regexp.Match("\"demand\":\\d", body)
	if err != nil || !matched {
		t.Errorf("Failed match in: '%s' (%v)", body, err)
	}
}
