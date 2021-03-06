// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"

	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/hosts"
)

type LocalClient struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	minimumTTL   int
	domainTTLMap map[string]uint32

	hosts   *hosts.Hosts
	rawName string
}

func NewLocalClient(q *dns.Msg, h *hosts.Hosts, minimumTTL int, domainTTLMap map[string]uint32) *LocalClient {
	c := &LocalClient{questionMessage: q.Copy(), hosts: h, minimumTTL: minimumTTL, domainTTLMap: domainTTLMap}
	c.rawName = c.questionMessage.Question[0].Name
	// require domain name is c.rawName
	return c
}

func (c *LocalClient) Exchange() *dns.Msg {
	if c.exchangeFromHosts() || c.exchangeFromIP() {
		if c.responseMessage != nil {
			common.SetMinimumTTL(c.responseMessage, uint32(c.minimumTTL))
			common.SetTTLByMap(c.responseMessage, c.domainTTLMap)
		}
		return c.responseMessage
	}

	return nil
}

func (c *LocalClient) exchangeFromHosts() bool {
	if c.hosts == nil {
		return false
	}

	name := c.rawName[:len(c.rawName)-1]
	ipv4List, ipv6List := c.hosts.Find(name)
	// regex ipv4 & ipv6 list
	if c.questionMessage.Question[0].Qtype == dns.TypeA && len(ipv4List) > 0 {
		var rrl []dns.RR
		for _, ip := range ipv4List {
			a, _ := dns.NewRR(c.rawName + " IN A " + ip.String())
			rrl = append(rrl, a)
		}
		c.setLocalResponseMessage(rrl)
		if c.responseMessage != nil {
			return true
		}
	} else if c.questionMessage.Question[0].Qtype == dns.TypeAAAA && len(ipv6List) > 0 {
		var rrl []dns.RR
		for _, ip := range ipv6List {
			aaaa, _ := dns.NewRR(c.rawName + " IN AAAA " + ip.String())
			rrl = append(rrl, aaaa)
		}
		c.setLocalResponseMessage(rrl)
		if c.responseMessage != nil {
			return true
		}
	}

	return false
}

func (c *LocalClient) exchangeFromIP() bool {
	name := c.rawName[:len(c.rawName)-1]
	ip := net.ParseIP(name)
	if ip == nil {
		return false
	}
	if ip.To4() == nil && ip.To16() != nil && c.questionMessage.Question[0].Qtype == dns.TypeAAAA {
		aaaa, _ := dns.NewRR(c.rawName + " IN AAAA " + ip.String())
		c.setLocalResponseMessage([]dns.RR{aaaa})
		return true
	} else if ip.To4() != nil && c.questionMessage.Question[0].Qtype == dns.TypeA {
		a, _ := dns.NewRR(c.rawName + " IN A " + ip.String())
		c.setLocalResponseMessage([]dns.RR{a})
		return true
	}

	return false
}

func (c *LocalClient) setLocalResponseMessage(rrl []dns.RR) {
	shuffleRRList := func(rrl []dns.RR) {
		rand.Seed(time.Now().UnixNano())
		for i := range rrl {
			j := rand.Intn(i + 1)
			rrl[i], rrl[j] = rrl[j], rrl[i]
		}
	}

	c.responseMessage = new(dns.Msg)
	for _, rr := range rrl {
		c.responseMessage.Answer = append(c.responseMessage.Answer, rr)
	}
	shuffleRRList(c.responseMessage.Answer)
	c.responseMessage.SetReply(c.questionMessage)
	c.responseMessage.RecursionAvailable = true
}
