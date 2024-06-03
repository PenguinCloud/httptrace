package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"
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
	avgGotCon         []float64
}

type httptracer interface {
	convertJSON(string) []byte
	getHttpTrace() *httptrace.ClientTrace
	findAvg() string
	writeJSON([]byte)
}

type totalInfo struct {
	h        hostInfo
	c        connInfo
	response *http.Response
	i        httptracer
}

func getFlags() (h hostInfo) {
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

func (t totalInfo) convertJSON(jstring string) []byte {
	var js fastjson.Parser
	w, _ := js.Parse(jstring)
	wf := w.MarshalTo(w.GetStringBytes())
	return wf
}

func (t totalInfo) writeJSON(js []byte) (err error) {
	err = os.WriteFile(fmt.Sprintf("httptrace-%s.json", t.h.host), js, 0600)
	return
}

// func (t totalInfo) getHttpTrace() (*httptrace.ClientTrace, *connoInf) {
func (t totalInfo) getHttpTrace() *httptrace.ClientTrace {
	c := t.c
	log.Println("Beginning Trace!")
	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			c.connStart = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			c.connEnd = time.Now()

			if info.Reused {
				log.Println("connection reused")
			} else {
				c.avgGotConn = append(c.avgGotConn, float64(c.connEnd.Sub(c.connStart).Microseconds()))

			}

		},
		ConnectStart: func(network, addr string) {
			c.connectStart = time.Now()

		},
		ConnectDone: func(network, addr string, err error) {
			c.connectEnd = time.Now()
			if err != nil {
				log.Println("error at ConnectDone", err)

			} else {
				c.avgConnect = append(c.avgConnect, float64(c.connectEnd.Sub(c.connectStart).Microseconds()))
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			c.dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			c.dnsEnd = time.Now()
			c.avgDns = append(c.avgDns, float64(c.dnsEnd.Sub(c.dnsStart).Microseconds()))

		},
		TLSHandshakeStart: func() {
			c.tlsHandShakeStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				log.Println("tls error", err)

			} else {
				c.tlsHandShakeEnd = time.Now()
				c.avgTlsHandShake = append(c.avgTlsHandShake, float64(c.tlsHandShakeEnd.Sub(c.tlsHandShakeStart).Microseconds()))

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
func (t totalInfo) findAvg() string {
	var js = "{"

	var (
		gotConn, connect, dns, tlsHandshake float64
	)
	for _, v := range t.c.avgGotConn {
		gotConn += v
	}
	js = js + fmt.Sprintf(`"AVG-EST": "%f",`, float64(gotConn)/float64(len(t.c.avgGotConn)))

	for _, v := range t.c.avgConnect {
		connect += v
	}
	js = js + fmt.Sprintf(`"AVG-CONN": "%f",`, float64(connect)/float64(len(t.c.avgConnect)))

	for _, v := range t.c.avgDns {
		dns += v
	}
	js = js + fmt.Sprintf(`"AVG-DNS": "%f",`, float64(dns)/float64(len(t.c.avgDns)))

	for _, v := range t.c.avgTlsHandShake {
		tlsHandshake += v
	}
	js = js + fmt.Sprintf(`"AVG-TLS-HS": "%f",`, float64(tlsHandshake)/float64(len(t.c.avgTlsHandShake)))
	js = js + "}"
	return js
}

func main() {
	var t totalInfo
	t.h = getFlags()
	cli := http.Client{}
	//var tome []byte
	url := "http://" + t.h.hostPort
	req, _ := http.NewRequest("GET", url, nil)
	//trace, tc := t.i.getHttpTrace()
	//t.c = *tc
	//req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	log.Println("Inializing Tracing Procedure... ")
	trace := t.i.getHttpTrace()
	log.Println("Inializing call... ")
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	_, _ = cli.Do(req)
	jstring := t.i.findAvg()
	bstring := t.i.convertJSON(jstring)
	t.i.writeJSON(bstring)
	return
}
