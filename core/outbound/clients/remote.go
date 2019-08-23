/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */
// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/common"
)

type RemoteClient struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	dnsUpstream *common.DNSUpstream
	inboundIP   string

	cache *cache.Cache
}

func NewClient(q *dns.Msg, u *common.DNSUpstream, ip string, cache *cache.Cache) *RemoteClient {
	c := &RemoteClient{questionMessage: q.Copy(), dnsUpstream: u, inboundIP: ip, cache: cache}

	return c
}

func (c *RemoteClient) Exchange(isLog bool) *dns.Msg {

	if c.responseMessage != nil {
		return c.responseMessage
	}

	var conn net.Conn
	if c.dnsUpstream.SOCKS5Address != "" {
		// If have sock5 proxy, dns will be transferred using socks5 proxy.
		s, err := proxy.SOCKS5(c.dnsUpstream.Protocol, c.dnsUpstream.SOCKS5Address, nil, proxy.Direct)
		if err != nil {
			log.Warnf("Failed to connect to SOCKS5 proxy: %s", err)
			return nil
		}
		conn, err = s.Dial(c.dnsUpstream.Protocol, c.dnsUpstream.Address)
		if err != nil {
			log.Warnf("Failed to connect to upstream via SOCKS5 proxy: %s", err)
			return nil
		}
	} else if c.dnsUpstream.Protocol == "tcp-tls" {
		// TCP-TLS DNS server
		var err error
		conf := &tls.Config{
			InsecureSkipVerify: false,
		}
		s := strings.Split(c.dnsUpstream.Address, "@")
		if len(s) == 2 {
			var servername, port string
			if servername, port, err = net.SplitHostPort(s[0]); err != nil {
				log.Warnf("Failed to parse DNS-over-TLS upstream address: %s", err)
				return nil
			}
			conf.ServerName = servername
			c.dnsUpstream.Address = net.JoinHostPort(s[1], port)
		}
		if conn, err = tls.Dial("tcp", c.dnsUpstream.Address, conf); err != nil {
			log.Warnf("Failed to connect to DNS-over-TLS upstream: %s", err)
			return nil
		}
	} else {
		// normal DNS server
		var err error
		if conn, err = net.Dial(c.dnsUpstream.Protocol, c.dnsUpstream.Address); err != nil {
			log.Warnf("Failed to connect to DNS upstream: %s", err)
			return nil
		}
	}
	// Time unit is second
	dnsTimeout := time.Duration(c.dnsUpstream.Timeout) * time.Second / 3

	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))

	dc := &dns.Conn{Conn: conn}
	defer dc.Close()
	err := dc.WriteMsg(c.questionMessage)
	// require dnsUpstream
	if err != nil {
		log.Warnf("%s Fail: Send question message failed", c.dnsUpstream.Name)
		return nil
	}
	temp, err := dc.ReadMsg()
	// read dnsUpstream response
	if err != nil {
		log.Debugf("%s Fail: %s", c.dnsUpstream.Name, err)
		return nil
	}
	if temp == nil {
		log.Debugf("Fail: Response message returned nil, maybe timeout? Please check your query or DNS configuration")
		return nil
	}

	c.responseMessage = temp

	if isLog {
		c.logAnswer("")
	}

	return c.responseMessage
}

func (c *RemoteClient) logAnswer(indicator string) {

	for _, a := range c.responseMessage.Answer {
		var name string
		// custom define log prefix
		if indicator != "" {
			name = indicator
		} else {
			name = c.dnsUpstream.Name
		}
		log.Debugf("Answer from %s: %s", name, a.String())
	}
}
