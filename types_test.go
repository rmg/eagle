package eagle

import (
	"encoding/xml"
	"testing"
)

func TestDeviceInfo(t *testing.T) {
	const in = `
  <DeviceInfo>
   <DeviceMacId>0xFFFFFFFFFFFFFFFF</DeviceMacId>
   <InstallCode>0xFFFFFFFFFFFFFFFF</InstallCode>
   <LinkKey>0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF</LinkKey>
   <FWVersion>{string}</FWVersion>
   <HWVersion>{string}</HWVersion>
   <ImageType>0xFFFF</ImageType>
   <Manufacturer>{string}</Manufacturer>
   <ModelId>{string}</ModelId>
   <DateCode>{string}</DateCode>
  </DeviceInfo>
  `
	out := DeviceInfoFragment{}
	err := xml.Unmarshal([]byte(in), &out)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
}

func TestUnmarshalRequest(t *testing.T) {
	const in = `<?xml version="1.0"?>
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
	out := InstantaneousDemand{}
	err := xml.Unmarshal([]byte(in), &out)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	if out.String() != "5.944kW" {
		t.Errorf("Got: '%s' instead of '5.944kW'", out)
	}
}

func TestPriceCluster(t *testing.T) {
	const in = `<?xml version="1.0"?>
  <rainforest macId="0xf0ad4e00ce69" timestamp="1355292588s">
  <PriceCluster>
    <DeviceMacId>0xd8d5b90000002aea</DeviceMacId>
    <MeterMacId>0x00078100007d67bb</MeterMacId>
    <TimeStamp>0x1b8d9cd0</TimeStamp>
    <Price>0x0000031d</Price>
    <Currency>0x007c</Currency>
    <TrailingDigits>0x04</TrailingDigits>
    <Tier>0x01</Tier>
    <StartTime>0xffffffff</StartTime>
    <Duration>0xffff</Duration>
    <RateLabel>Block 1</RateLabel>
  </PriceCluster>
  </rainforest>
  `
	out := PriceCluster{}
	err := xml.Unmarshal([]byte(in), &out)
	if err != nil {
		t.Errorf("error: %+v", err)
		return
	}
	if out.String() != "0.797 CAD" {
		t.Errorf("Got: '%s' instead of '0.797 CAD'", out)
	}
}

func TestFragmentNamer(t *testing.T) {
	const priceCluster = `<?xml version="1.0"?>
  <rainforest macId="0xf0ad4e00ce69" timestamp="1355292588s">
  <PriceCluster></PriceCluster>
  </rainforest>
  `
	const message = `<?xml version="1.0"?>
  <rainforest macId="0xf0ad4e00ce69" timestamp="1355292588s">
  <Message></Message>
  </rainforest>
  `
	tests := []struct {
		input    string
		expected string
	}{
		{
			priceCluster, "PriceCluster"},
		{message, "Message"},
	}
	for _, test := range tests {
		out := Request{}
		err := xml.Unmarshal([]byte(test.input), &out)
		if err != nil {
			t.Errorf("error: %+v", err)
		}
		if out.Fragment.XMLName.Local != test.expected {
			t.Errorf("%s expected, got: %+v", test.expected, out.Fragment.XMLName)
		}
	}
}
