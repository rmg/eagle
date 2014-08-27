package eagle

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

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
		metrics = append(metrics, result)
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
		metrics = append(metrics, result)
	}
}
