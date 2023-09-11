package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

var pollInterval int
var reportInterval int
var addr string

var metrics = map[string]bool{
	"Alloc":         true,
	"BuckHashSys":   true,
	"Frees":         true,
	"GCCPUFraction": true,
	"GCSys":         true,
	"HeapAlloc":     true,
	"HeapIdle":      true,
	"HeapInuse":     true,
	"HeapObjects":   true,
	"HeapReleased":  true,
	"HeapSys":       true,
	"LastGC":        true,
	"Lookups":       true,
	"MCacheInuse":   true,
	"MCacheSys":     true,
	"MSpanInuse":    true,
	"MSpanSys":      true,
	"Mallocs":       true,
	"NextGC":        true,
	"NumForcedGC":   true,
	"NumGC":         true,
	"OtherSys":      true,
	"PauseTotalNs":  true,
	"StackInuse":    true,
	"StackSys":      true,
	"Sys":           true,
	"TotalAlloc":    true,
	"PollCount":     true,
	"RandomValue":   true,
}

var valuesGauge = map[string]float64{}
var pollCount uint64

func main() {

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	go getMetrics()

	time.Sleep(time.Duration(reportInterval) * time.Second)

	for {
		for k, v := range valuesGauge {
			post("gauge", k, strconv.FormatFloat(v, 'f', -1, 64))
		}
		fmt.Println(pollCount)
		post("counter", "PollCount", strconv.FormatUint(pollCount, 10))
		post("gauge", "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
		pollCount = 0
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func getMetrics() {
	var rtm runtime.MemStats

	for {
		pollCount += 1
		runtime.ReadMemStats(&rtm)
		numfield := reflect.ValueOf(&rtm).Elem().NumField()
		for x := 0; x < numfield; x++ {
			metricsName := reflect.TypeOf(&rtm).Elem().Field(x).Name
			if metrics[metricsName] {
				metricsValue := reflect.ValueOf(&rtm).Elem().Field(x)
				var metricsFloat float64
				if metricsValue.CanFloat() {
					metricsFloat = float64(metricsValue.Float())
				} else if metricsValue.CanUint() {
					metricsFloat = float64(metricsValue.Uint())
				}
				println(metricsName, metricsFloat)
				valuesGauge[metricsName] = metricsFloat
			}
		}
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func post(mType string, mName string, mValue string) {
	resp, err := http.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", addr, mType, mName, mValue), "text/plain", bytes.NewReader([]byte{}))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
