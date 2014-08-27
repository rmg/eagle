package eagle

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/rmg/iso4217"
	"math"
	"net"
	"strconv"
)

type HexInt int64

func (i *HexInt) UnmarshalText(b []byte) error {
	newInt, err := strconv.ParseInt(string(b), 0, 64)
	*i = HexInt(newInt)
	return err
}

type YNBool bool

func (v *YNBool) UnmarshalText(b []byte) error {
	switch b[0] {
	case 'Y', 'y':
		*v = true
	case 'N', 'n':
		*v = false
	default:
		*v = false
	}
	return nil
}

type MacAddrHex net.HardwareAddr

func (m MacAddrHex) String() string {
	return net.HardwareAddr(m).String()
}

func (m *MacAddrHex) UnmarshalText(b []byte) error {
	bytes, err := hex.DecodeString(string(b[2:]))
	*m = MacAddrHex(bytes)
	return err
}

type RainforestDocument struct {
	MacId     MacAddrHex `xml:"macId,attr"`
	Version   string     `xml:"version,attr"`
	Timestamp string     `xml:"timestamp,attr"`
}

type RequestFragment struct {
	XMLName xml.Name
}

type Request struct {
	RainforestDocument
	Fragment RequestFragment `xml:",any"`
}

type DeviceInfoFragment struct {
	XMLName      string // `xml:"rainforest>*"`
	DeviceMacId  MacAddrHex
	InstallCode  string // 16 hex digits Install Code for EAGLE™ ZigBee radio
	LinkKey      string // 32 hex digits ZigBee radio Link Key
	FWVersion    string // Text Firmware Version
	HWVersion    string // Text Hardware Version
	ImageType    string // 4 hex digits ZigBee code image type
	Manufacturer string // Text “Rainforest Automation”
	ModelId      string // Text “RFA-Z109”
	DateCode     string // YYYYMMDDZZZZZZZZ Manufacturer’s date code and lot number
	// <Port>/dev/ttySP0</Port>
}

type DeviceInfo struct {
	RainforestDocument
	DeviceInfo DeviceInfoFragment
}

type NetworkInfoFragment struct {
	DeviceMacId MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	CoordMacId  string     // 16 hex digits MAC Address of Meter
	Status      string
	// Initializing | Network
	// Discovery | Joining | Join: Fail
	// | Join: Success |
	// Authenticating |
	// Authenticating: Success |
	// Authenticating: Fail |
	// Connected | Disconnected |
	// Rejoining
	//Indicates the current state of the EAGLE™ ZigBee radio.
	Description  string // Text; Optional Description of ZigBee radio.state
	StatusCode   string // 2 hex digits; Optional Status code for the current state
	ExtPanId     string // 16 hex digits; Optional Extended PAN ID of the ZigBee network
	Channel      string // 11 – 26; Optional Indicates the radio channel on which the EAGLE™ is operating
	ShortAddr    string // 4 hex digits; Optional The short address assigned to the EAGLE™ by the network coordinator
	LinkStrength string // 0x00 – 0x64 Indicates the strength of the radio link
}

type NetworkInfo struct {
	RainforestDocument
	NetworkInfo NetworkInfoFragment
}

type InstantaneousDemandFragment struct {
	XMLName             xml.Name
	DeviceMacId         MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId          string     // 16 hex digits MAC Address of Meter
	TimeStamp           string     // Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) when demand data was received from meter.
	Demand              HexInt     // 6 hex digits The raw instantaneous demand value. This is a 24-bit signed integer.
	Multiplier          HexInt     // Up to 8 hex digits The multiplier; if zero, use 1
	Divisor             HexInt     // Up to 8 hex digits The divisor; if zero, use 1
	DigitsRight         HexInt     // Up to 2 hex digits Number of digits to the right of the decimal point to display
	DigitsLeft          HexInt     // Up to 2 hex digits Number of digits to the left of the decimal point to display
	SuppressLeadingZero YNBool     // Y | N Y: Do not display leading zeros, N: Display leading zeros
}

type InstantaneousDemand struct {
	RainforestDocument
	InstantaneousDemand InstantaneousDemandFragment
}

func (i InstantaneousDemand) String() string {
	d := i.InstantaneousDemand
	demand := float64(d.Demand)
	mult := float64(d.Multiplier)
	div := float64(d.Divisor)
	if div == 0 {
		div = 1
	}
	if mult == 0 {
		mult = 1
	}
	format := fmt.Sprintf("%%%d.%dfkW", d.DigitsLeft, d.DigitsRight)
	return fmt.Sprintf(format, demand*mult/div)
}

func (i InstantaneousDemand) Int() int {
	return int(i.InstantaneousDemand.Demand)
}

// <PriceCluster>
//   <DeviceMacId>0xd8d5b90000002aea</DeviceMacId>
//   <MeterMacId>0x00078100007d67bb</MeterMacId>
//   <TimeStamp>0x1b8d9cd0</TimeStamp>
//   <Price>0x0000031d</Price>
//   <Currency>0x007c</Currency>
//   <TrailingDigits>0x04</TrailingDigits>
//   <Tier>0x01</Tier>
//   <StartTime>0xffffffff</StartTime>
//   <Duration>0xffff</Duration>
//   <RateLabel>Block 1</RateLabel>
// </PriceCluster>
type PriceClusterFragment struct {
	DeviceMacId    MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId     MacAddrHex // 16 hex digits MAC Address of Meter
	TimeStamp      HexInt     // Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) when price data was received from meter or set by user
	Price          HexInt     // Up to 8 hex digits Price from meter or set by user; will be zero if no price is set
	Currency       HexInt     // Up to 4 hex digits Currency being used; value of this field matches the values defined by ISO 4217
	TrailingDigits HexInt     // Up to 2 hex digits The number of implicit decimal places in the price. (e.g. 2 means divide Price by 100).
	Tier           string     // 1 - 5 The price Tier in effect.
	//   <StartTime>0xffffffff</StartTime>
	//   <Duration>0xffff</Duration>
	RateLabel string // Text Rate label for the current price tier; will be “Set by User” if a user-defined price is set
}

type PriceCluster struct {
	RainforestDocument
	PriceCluster PriceClusterFragment
}

func (p PriceCluster) String() string {
	name, _ := iso4217.ByCode(int(p.PriceCluster.Currency))
	price := int(p.PriceCluster.Price)
	div := int(math.Pow10(int(p.PriceCluster.TrailingDigits)))
	if div == 1 {
		return fmt.Sprintf("%d %s", price, name)
	} else {
		return fmt.Sprintf("%d.%d %s", price/div, price%div, name)
	}
}

func (p PriceCluster) Int() int {
	return int(p.PriceCluster.Price)
}

type MessageFragment struct {
	DeviceMacId          MacAddrHex //  16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId           string     //  16 hex digits MAC Address of Meter
	TimeStamp            string     //  Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) when message was received from meter
	Id                   string     //  Up to 8 hex digits Message ID from meter
	Text                 string     //  Text Contents of message, HTML encoded: &gt; replaces the > character &lt; replaces the < character &amp; replaces the & character &quot; replaces the " character
	Priority             string     //  Low | Medium | High | Critical Message priority
	ConfirmationRequired string     //  Y | N Y: a user confirmation is required; N: a user confirmation is not required (default)
	Confirmed            string     //  Y | N Y: the user confirmation has been sent; N: the user confirmation has not been sent (default)
	Queue                string     //  Active | Cancel Pending Active: Indicates message is in active queue Cancel Pending: Indicates message is in cancel pending queue
}

type Message struct {
	RainforestDocument
	Message MessageFragment
}

type CurrentSummationFragment struct {
	DeviceMacId         MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId          string     // 16 hex digits MAC Address of Meter
	TimeStamp           string     // Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) when demand data was received from meter.
	SummationDelivered  string     // Up to 8 hex digitsThe raw value of the total summation of commodity delivered from the utility to the user.
	SummationReceived   string     // Up to 8 hex digits The raw value of the total summation of commodity received from the user by the utility.
	Multiplier          string     // Up to 8 hex digits The multiplier; if zero, use 1
	Divisor             string     // Up to 8 hex digits The divisor; if zero, use 1
	DigitsRight         string     // Up to 2 hex digits Number of digits to the right of the decimal point to display
	DigitsLeft          string     // Up to 2 hex digits Number of digits to the left of the decimal point to display
	SuppressLeadingZero string     // Y | N Y: Do not display leading zeros N: Display leading zeros
}

type CurrentSummation struct {
	RainforestDocument
	CurrentSummation CurrentSummationFragment
}

type MeterInfoFragment struct {
	DeviceMacId MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId  string     // 16 hex digits MAC Address of Meter
	Type        string     // electric | gas | water | other Type of meter
	Nickname    string     // Text Nickname set for the meter
	Account     string     // Text Account Identification
	Auth        string     // Text Authentication code
	Host        string     // Text Hosting Provider
	Enabled     string     // Y | N Y: to start transmitting data to host N: to stop transmitting data to host
}

type MeterInfo struct {
	RainforestDocument
	MeterInfo MeterInfoFragment
}

type FastPollStatusFragment struct {
	DeviceMacId MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId  string     // 16 hex digits MAC Address of Meter
	Frequency   string     // 0x01 – 0xFF Frequency to poll meter, in seconds
	EndTime     string     // Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) when fast poll will end. If EndTime is earlier than the current time, then fast poll mode is turned off.
}

type FastPollStatus struct {
	RainforestDocument
	FastPollStatus FastPollStatusFragment
}

type HistoryDataFragment struct {
	CurrentSummation []CurrentSummationFragment // `xml:"HistoryData>CurrentSummation"`
}

type HistoryData struct {
	RainforestDocument
	HistoryData HistoryDataFragment
}

type ProfileDataFragment struct {
	DeviceMacId MacAddrHex // 16 hex digits MAC Address of EAGLE™ ZigBee radio
	MeterMacId  string     // 16 hex digits MAC Address of Meter
	EndTime     string     // Up to 8 hex digits UTC Time (offset in seconds from 00:00:00 01Jan2000) of the end of the most chronologically recent interval; 0x0 indicates the most recent interval block.
	Status      string     //  0x0 – 0x05 Status of returned data:
	// 0x00 Success
	// 0x01 Undefined Interval Channel requested
	// 0x02 Interval Channel not supported
	// 0x03 Invalid End Time
	// 0x04 More periods Requested than can be returned
	// 0x05 No intervals available for the requested time
	ProfileIntervalPeriod string // 0 - 7
	// The length of each sampling interval:
	// 0 Daily
	// 1 60 minutes
	// 2 30 minutes
	// 3 15 minutes
	// 4 10 minutes
	// 5 7.5 minutes
	// 6 5 minutes
	// 7 2.5 minutes
	NumberOfPeriodsDelivered string // 0x0 – 0xFF The number of intervals being returned.
	IntervalDataX            string // Up to 6 hex digits
	// X = 1-12. Series of up to 12 interval data points from the meter. Most recent interval is first; oldest is last. Invalid intervals are marked as 0xFFFFFF.
}

type ProfileData struct {
	RainforestDocument
	ProfileData ProfileDataFragment
}

// <MessageCluster>
//   <DeviceMacId>0xd8d5b90000002aea</DeviceMacId>
//   <MeterMacId>0x00078100007d67bb</MeterMacId>
//   <TimeStamp>0x1b8d9bf3</TimeStamp>
//   <Id>0x000001b1</Id>
//   <Text>Registration Successful</Text>
//   <Priority>Medium</Priority>
//   <StartTime>0x1b878fa1</StartTime>
//   <Duration>0x545e</Duration>
//   <ConfirmationRequired>Y</ConfirmationRequired>
//   <Confirmed>Y</Confirmed>
//   <Queue>Active</Queue>
// </MessageCluster>

// <BlockPriceDetail>
//   <DeviceMacId>0xd8d5b90000002aea</DeviceMacId>
//   <MeterMacId>0x00078100007d67bb</MeterMacId>
//   <TimeStamp>0x1b8d9c25</TimeStamp>
//   <CurrentStart>0x1b89a6a2</CurrentStart>
//   <CurrentDuration>0x2ad1</CurrentDuration>
//   <BlockPeriodConsumption>0x0000000000038fa2</BlockPeriodConsumption>
//   <BlockPeriodConsumptionMultiplier>0x00000001</BlockPeriodConsumptionMultiplier>
//   <BlockPeriodConsumptionDivisor>0x000003e8</BlockPeriodConsumptionDivisor>
//   <NumberOfBlocks>0x02</NumberOfBlocks>
//   <Multiplier>0x00000001</Multiplier>
//   <Divisor>0x00000001</Divisor>
//   <Currency>0x007c</Currency>
//   <TrailingDigits>0x04</TrailingDigits>
//   <Price1>0x0000031d</Price1>
//   <Threshold1>0x000002c6</Threshold1>
//   <Price2>0x000004ab</Price2>
// </BlockPriceDetail>

type Command struct {
	Name  string
	MacId MacAddrHex
}

type RavenCommand struct {
	Command
	// set_fast_poll
	Frequency HexInt
	Duration  HexInt
	// get_profile_data
	DeviceMacId     MacAddrHex
	MeterMacId      MacAddrHex
	NumberOfPeriods HexInt
	EndTime         HexInt
	IntervalChannel string
	// set_schedule
	// DeviceMacId MacAddrHex
	Event string
	// Frequency HexInt
	Enabled string
}

type LocalCommand struct {
	Command
	// get_fast_poll_status
	// ...
	// get_history_data
	StartTime HexInt
	EndTime   HexInt
	Frequency HexInt
	//
}

type SetFastPoll struct {
}
