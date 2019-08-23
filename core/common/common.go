// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package common provides common functions.
package common

import (
	"net"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var ReservedIPNetworkList = getReservedIPNetworkList()

func IsIPMatchList(ip net.IP, ipNetList []*net.IPNet, isLog bool, name string) bool {
	if ipNetList != nil {
		for _, ipNet := range ipNetList {
			if ipNet.Contains(ip) {
				if isLog {
					log.Debugf("Matched: IP network %s %s %s", name, ip.String(), ipNet.String())
				}
				return true
			}
		}
	} else {
		log.Debug("IP network list is nil, not checking")
	}

	return false
}

func IsDomainMatchRule(pattern string, domain string) bool {
	matched, err := regexp.MatchString(pattern, domain)
	if err != nil {
		log.Warnf("Error matching domain %s with pattern %s: %s", domain, pattern, err)
	}
	return matched
}

func HasAnswer(m *dns.Msg) bool { return m != nil && len(m.Answer) != 0 }

func HasSubDomain(s string, sub string) bool {
	return strings.HasSuffix(sub, "."+s) || s == sub
}

func getReservedIPNetworkList() []*net.IPNet {
	var ipNetList []*net.IPNet
	localCIDR := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10"}
	for _, c := range localCIDR {
		_, ipNet, err := net.ParseCIDR(c)
		if err != nil {
			break
		}
		ipNetList = append(ipNetList, ipNet)
	}
	return ipNetList
}

func FindRecordByType(msg *dns.Msg, t uint16) string {
	for _, rr := range msg.Answer {
		if rr.Header().Rrtype == t {
			items := strings.SplitN(rr.String(), "\t", 5)
			return items[4]
		}
	}

	return ""
}

func SetMinimumTTL(msg *dns.Msg, minimumTTL uint32) {
	if minimumTTL == 0 {
		return
	}
	for _, a := range msg.Answer {
		if a.Header().Ttl < minimumTTL {
			a.Header().Ttl = minimumTTL
		}
	}
}

func SetTTLByMap(msg *dns.Msg, domainTTLMap map[string]uint32) {
	if len(domainTTLMap) == 0 {
		return
	}
	for _, a := range msg.Answer {
		name := a.Header().Name[:len(a.Header().Name)-1]
		for k, v := range domainTTLMap {
			if IsDomainMatchRule(k, name) {
				a.Header().Ttl = v
			}
		}
	}
}
