package traffic_control

import (
	"http-server.org/gazer/tools"
	"sync"
	"time"
)

var trafficStat *tools.Statistics
var mtx sync.Mutex
var stopping bool
var lastReceivedBytes int
var lastSentBytes int
var lastStatDT time.Time
var actualSnd float64
var actualRcv float64

func Start() {
	trafficStat = tools.NewStatistics()
	go thWorker()
}

func Stop() {
	stopping = true
}

func Stat() *tools.Statistics {
	return trafficStat
}

func thWorker() {
	for !stopping {
		time.Sleep(1000 * time.Millisecond)

		mtx.Lock()
		t := time.Now().UTC()
		duration := t.Sub(lastStatDT).Seconds()
		actualSnd = float64(0)
		actualRcv = float64(0)
		if duration > 0 {
			sent := trafficStat.Get("snd")
			received := trafficStat.Get("rcv")
			actualSnd = float64(sent-lastSentBytes) / duration
			actualRcv = float64(received-lastReceivedBytes) / duration
			lastStatDT = time.Now().UTC()
			lastSentBytes = sent
			lastReceivedBytes = received
		}
		mtx.Unlock()
	}
}

func AddSend(value int) {
	trafficStat.Add("snd", value)
}

func GetSend() float64 {
	mtx.Lock()
	defer mtx.Unlock()
	return actualSnd
}

func AddRcv(value int) {
	trafficStat.Add("rcv", value)
}

func GetRcv() float64 {
	mtx.Lock()
	defer mtx.Unlock()
	return actualRcv
}
