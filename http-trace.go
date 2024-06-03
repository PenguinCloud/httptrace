package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"flag"
	"fastjson"
)


var hostInfo struct {
	host *string
	port *string
	tls *Bool
	hostPort *string
}

var connInfo struct {
	dnsStart string
	dnsEnd string
	connStart string
	connEnd string
	connectStart string
	connectEnd string
	tlsHandShakeStart string
	tlsHandShakeEnd time.Time
	avgGotConn string
	avgConnect string
	avgDns string
}

func (h hostInfo, t tracker) main() {
	h.host := flag.String("host", "127.0.0.1", "The target host")
	h.port := flag.String("port", 80, "The target port")
	h.tls := flag.Bool("tls", false, "Use TLS")
	h.hostPort = h.host + ":" + h.port
	flag.Parse()
	cli := http.Client{}
	req, _ := http.NewRequest("GET", "http://"+hostPtr+":"+portPtr, bytes.NewBuffer([]byte))
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), getHttpTrace(&h)))
 	response,err:= cli.Do(req)
}

func (h hostInfo) convertJSON() *fastjson.Parse {

}

func (c connInfo) getHttpTrace(*h hostInfo) (*httptrace.ClientTrace) {
   
	trace := &httptrace.ClientTrace{
	 GetConn: func(h.host string) {
	  c.connStart = time.Now()
	 },
	 GotConn: func(info httptrace.GotConnInfo) {
	  c.connEnd = time.Now()
   
	  if info.Reused {
	   log.Println("connection reused")
	  } else {
	   c.avgGotConn = append(c.avgGotConn, c.connEnd.Sub(c.connStart).Microseconds())
   
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
	   c.avgConnect = append(c.avgConnect, c.connectEnd.Sub(c.connectStart).Microseconds())
	  }
	 },
	 DNSStart: func(info httptrace.DNSStartInfo) {
	  c.dnsStart = time.Now()
	 },
	 DNSDone: func(info httptrace.DNSDoneInfo) {
	  c.dnsEnd = time.Now()
	  c.avgDns = append(avgDns, c.dnsEnd.Sub(c.dnsStart).Microseconds())
   
	 },
	 TLSHandshakeStart: func() {
	  c.tlsHandShakeStart = time.Now()
	 },
	 TLSHandshakeDone: func(state tls.ConnectionState, err error) {
	  if err != nil {
	   log.Println("tls error", err)
   
	  } else {
	   c.tlsHandShakeEnd = time.Now()
	   c.avgTlsHandShake = append(avgTlsHandShake, tlsHandShakeEnd.Sub(tlsHandShakeStart).Microseconds())
   
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
   func findAvg(*c connInfo) {
	log.Println("{")
	var (
	 gotConn, connect, dns, tlsHandshake int64
	)
	for _, v := range avgGotConn {
	 gotConn += v
	}
	log.Println("AVG-EST", float64(gotConn)/float64(len(avgGotConn)))
   
	for _, v := range avgConnect {
	 connect += v
	}
	log.Println("AVG-CONN", float64(connect)/float64(len(avgConnect)))
   
	for _, v := range avgDns {
	 dns += v
	}
	log.Println("AVG-DNS", float64(dns)/float64(len(avgDns)))
   
	for _, v := range avgTlsHandShake {
	 tlsHandshake += v
	}
	log.Println("AVG-TLS-HS", float64(tlsHandshake)/float64(len(avgTlsHandShake)))
   }