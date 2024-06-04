package main

// https://github.com/gocolly/colly/blob/master/http_trace.go
import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"

	"github.com/valyala/fastjson"
)

type hostInfo struct {
	host     string
	port     string
	tls      bool
	hostPort string
}

type connInfo struct {
	dnsStart          time.Time
	dnsEnd            time.Time
	connStart         time.Time
	connEnd           time.Time
	connectStart      time.Time
	connectEnd        time.Time
	tlsHandShakeStart time.Time
	tlsHandShakeEnd   time.Time
	avgGotConn        []float64
	avgConnect        []float64
	avgDns            []float64
	avgTlsHandShake   []float64
	//avgGotCon         []float64
}

type jstruct struct {
	host string `json:"host"`
	port string `json:"port"`
	avgEST string `json:"avgEST"`
	avgCONN string `json:"avgCONN"`
	avgDNS string `json:"avgDns"`
	avgTLSHS string `json: "avgTLSHS"`
}

var t totalInfo

type httptracer interface {
	convertJSON(string) []byte
	findAvg() string
	writeJSON([]byte)
}

type totalInfo struct {
	h        hostInfo
	c        connInfo
	i        httptracer
	j 	     jstruct
}

func getFlags() () {
	var h = t.h
	log.Println("Checking for flags...")
	h.host = *flag.String("host", "127.0.0.1", "The target host")
	h.port = *flag.String("port", "8000", "The target port")
	h.tls = *flag.Bool("tls", false, "Use TLS")
	var joiner []string
	joiner = append(joiner, h.host)
	joiner = append(joiner, ":")
	h.hostPort = strings.Join(joiner, h.port)
	flag.Parse()
	return
}

func convertJSON(jstring string) []byte {
	w, _ := fastjson.Parse(jstring)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(jstring ), &jsonMap)
	return jsonMap
}

func writeJSON(js []byte) (err error) {
	err = os.WriteFile(fmt.Sprintf("httptrace-%s.json", t.h.host), js, 0600)
	return
}

// func (t totalInfo) getHttpTrace() (*httptrace.ClientTrace, *connoInf) {
func getHttpTrace() *httptrace.ClientTrace {
	var t totalInfo
	log.Println("Beginning Trace!")
	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			t.c.connStart = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			t.c.connEnd = time.Now()

			if info.Reused {
				log.Println("connection reused")
			} else {
				t.c.avgGotConn = append(t.c.avgGotConn, float64(t.c.connEnd.Sub(t.c.connStart).Microseconds()))

			}

		},
		ConnectStart: func(network, addr string) {
			t.c.connectStart = time.Now()

		},
		ConnectDone: func(network, addr string, err error) {
			t.c.connectEnd = time.Now()
			if err != nil {
				log.Println("error at ConnectDone", err)

			} else {
				t.c.avgConnect = append(t.c.avgConnect, float64(t.c.connectEnd.Sub(t.c.connectStart).Microseconds()))
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			t.c.dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t.c.dnsEnd = time.Now()
			t.c.avgDns = append(t.c.avgDns, float64(t.c.dnsEnd.Sub(t.c.dnsStart).Microseconds()))

		},
		TLSHandshakeStart: func() {
			t.c.tlsHandShakeStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				log.Println("tls error", err)

			} else {
				t.c.tlsHandShakeEnd = time.Now()
				t.c.avgTlsHandShake = append(t.c.avgTlsHandShake, float64(t.c.tlsHandShakeEnd.Sub(t.c.tlsHandShakeStart).Microseconds()))

			}

		},
		PutIdleConn: func(err error) {
			if err != nil {
				log.Println("error at putIdleConn", err)
			} else {
				log.Println("put idle connection")
			}

		},
	}

	return trace

}

// finding average of each operation
func findAvg() string {
	var js = "{"

	var (
		gotConn, connect, dns, tlsHandshake float64
	)
	for _, v := range t.c.avgGotConn {
		gotConn += v
	}
	t.j.avgCONN = float64(gotConn)/float64(len(t.c.avgGotConn)))

	for _, v := range t.c.avgConnect {
		connect += v
	}
	js = js + fmt.Sprintf(`"avgCONN": "%f",`, float64(connect)/float64(len(t.c.avgConnect)))

	for _, v := range t.c.avgDns {
		dns += v
	}
	js = js + fmt.Sprintf(`"avgDNS": "%f",`, float64(dns)/float64(len(t.c.avgDns)))

	for _, v := range t.c.avgTlsHandShake {
		tlsHandshake += v
	}
	js = js + fmt.Sprintf(`"avgTLSHS": "%f",`, float64(tlsHandshake)/float64(len(t.c.avgTlsHandShake)))
	js = js + "}"
	return js
}

func main() {
	getFlags()
	cli := http.Client{}
	//var tome []byte
	url := "http://" + t.h.hostPort
	req, _ := http.NewRequest("GET", url, nil)
	//trace, tc := t.i.getHttpTrace()
	//t.c = *tc
	//req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	log.Println("Inializing Tracing Procedure... ")
	trace := getHttpTrace()
	log.Println("Inializing call... ")
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	_, _ = cli.Do(req)
	jstring := findAvg()
	bstring := convertJSON(jstring)
	writeJSON(bstring)
}
