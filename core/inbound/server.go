// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package inbound implements dns server for inbound connection.
package inbound

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/import-yuefeng/smartDNS/core/outbound"
)

type Server struct {
	bindAddress      string
	debugHttpAddress string
	dispatcher       outbound.Dispatcher
	rejectQType      []uint16
}

// NewServer func create new Server struct object
func NewServer(bindAddress string, debugHTTPAddress string, dispatcher outbound.Dispatcher, rejectQType []uint16) *Server {
	return &Server{
		bindAddress:      bindAddress,
		debugHttpAddress: debugHTTPAddress,
		dispatcher:       dispatcher,
		rejectQType:      rejectQType,
	}
}

// DumpCache func be used debug
func (s *Server) DumpCache(w http.ResponseWriter, req *http.Request) {
	if s.dispatcher.Cache == nil {
		io.WriteString(w, "error: cache not enabled")
		return
	}

	type answer struct {
		Name  string `json:"name"`
		TTL   int    `json:"ttl"`
		Type  string `json:"type"`
		Rdata string `json:"rdata"`
	}

	type response struct {
		Length   int                  `json:"length"`
		Capacity int                  `json:"capacity"`
		Body     map[string][]*answer `json:"body"`
	}

	query := req.URL.Query()
	nobody := true
	if t := query.Get("nobody"); strings.ToLower(t) == "false" {
		nobody = false
	}

	rs, l := s.dispatcher.Cache.Dump(nobody)
	body := make(map[string][]*answer)

	for k, es := range rs {
		var answers []*answer
		for _, e := range es {
			ts := strings.Split(e, "\t")
			ttl, _ := strconv.Atoi(ts[1])
			r := &answer{
				Name:  ts[0],
				TTL:   ttl,
				Type:  ts[3],
				Rdata: ts[4],
			}
			answers = append(answers, r)
		}
		body[strings.TrimSpace(k)] = answers
	}

	res := response{
		Body:     body,
		Length:   l,
		Capacity: s.dispatcher.Cache.Capacity(),
	}

	responseBytes, err := json.Marshal(&res)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	io.WriteString(w, string(responseBytes))
}

//Run func bind smartDNS listen port and address
func (s *Server) Run() {
	mux := dns.NewServeMux()
	mux.Handle(".", s)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	log.Infof("smartDNS is listening on %s", s.bindAddress)

	for _, p := range [2]string{"tcp", "udp"} {
		go func(p string) {
			err := dns.ListenAndServe(s.bindAddress, p, mux)
			if err != nil {
				log.Fatalf("Listening on port %s failed: %s", p, err)
				os.Exit(1)
			}
		}(p)
	}

	if s.debugHttpAddress != "" {
		http.HandleFunc("/cache", s.DumpCache)
		wg.Add(1)
		go http.ListenAndServe(s.debugHttpAddress, nil)
	}
	// Process will obstruct procesing require
	wg.Wait()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {
	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	// require ip addr
	log.Debugf("Question from %s: %s", inboundIP, q.Question[0].String())

	for _, qt := range s.rejectQType {
		if isQuestionType(q, qt) {
			return
		}
	}

	responseMessage := s.dispatcher.Exchange(q, inboundIP)

	if responseMessage == nil {
		dns.HandleFailed(w, q)
		return
	}

	err := w.WriteMsg(responseMessage)
	if err != nil {
		log.Warnf("Write message failed, message: %s, error: %s", responseMessage, err)
		return
	}
}

func isQuestionType(q *dns.Msg, qt uint16) bool { return q.Question[0].Qtype == qt }
