package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func graphiteMetric(name string, value int) {
	strTpl := "%s.%s %d\n"
	apiKey := os.Getenv("HOSTEDGRAPHITE_APIKEY")
	datum := fmt.Sprintf(strTpl, apiKey, name, value)
	conn, _ := net.Dial("udp", "carbon.hostedgraphite.com:2003")
	defer conn.Close()
	conn.Write([]byte(datum))
}

func forwardMetric(name string, value int) {
	url := "http://sandbox.influxdb.com:8086/db/eagle/series?u=rmg&p=GpqW1DtL3Png"
	jsonTpl := "[{ \"name\": \"%s\", \"columns\": [\"value\"], \"points\": [[%d]] }]"
	jsonStr := fmt.Sprintf(jsonTpl, name, value)

	req, _ := http.NewRequest("POST", url, strings.NewReader(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	// fmt.Println("request Body:", jsonStr)
	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
}

func MetricsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		ReceiveMetrics(w, req)
	} else if req.Method == "GET" {
		ReportMetrics(w, req)
	} else {
		log.Printf("Received bad request: %s %s", req.Method, req.URL)
		w.WriteHeader(400)
	}
}

func ReportMetrics(w http.ResponseWriter, req *http.Request) {
	res, err := json.Marshal(metrics)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Error: %+v", err)))
	} else {
		w.Header().Set(http.CanonicalHeaderKey("content-type"), "application/json")
		w.Write(res)
	}
}

type Reading struct {
	Time   time.Time `json:"time"`
	Demand int       `json:"demand"`
	Price  int       `json:"price"`
}

var metrics = make([]Reading, 1)

func ReceiveMetrics(w http.ResponseWriter, req *http.Request) {
	msg := Request{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		log.Printf("500 from %v: %s\n", req, err)
	}
	err = xml.Unmarshal(body, &msg)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		log.Printf("500 from %+v: %s\n", req, err)
	} else {
		reqType := msg.Fragment.XMLName.Local
		switch reqType {
		case "InstantaneousDemand":
			ReceiveDemand(w, req, body)
		case "PriceCluster":
			ReceivePrice(w, req, body)
		default:
			w.WriteHeader(200)
			log.Printf("%s", reqType)
		}

	}
}

func ReceiveDemand(w http.ResponseWriter, req *http.Request, body []byte) {
	demand := InstantaneousDemand{}
	err := xml.Unmarshal(body, &demand)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("500 from %+v: %v", req, err)
	} else {
		result := Reading{time.Now(), demand.Int(), metrics[len(metrics)-1].Price}
		log.Printf("InstantaneousDemand: %+v", result)
		forwardMetric("demand", demand.Int())
		graphiteMetric("demand", demand.Int())
		// metrics = append(metrics, result)
	}
}

func ReceivePrice(w http.ResponseWriter, req *http.Request, body []byte) {
	price := PriceCluster{}
	err := xml.Unmarshal(body, &price)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("500 from %+v: %v", req, err)
	} else {
		result := Reading{time.Now(), metrics[len(metrics)-1].Demand, price.Int()}
		log.Printf("PriceCluster: %+v", result)
		forwardMetric("price", price.Int())
		graphiteMetric("price", price.Int())
		// metrics = append(metrics, result)
	}
}
