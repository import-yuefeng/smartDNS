// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package outbound

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/hosts"
	"github.com/import-yuefeng/smartDNS/core/matcher"
	"github.com/import-yuefeng/smartDNS/core/outbound/clients"
)

type Dispatcher struct {
	RedirectIPv6Record bool
	MinimumTTL         int
	DomainTTLMap       map[string]uint32
	DefaultDNSBundle   string
	DNSFilter          map[string]*common.Filter
	DNSBunch           map[string][]*common.DNSUpstream
	Hosts              *hosts.Hosts
	Cache              *cache.Cache
}

type HitTask struct {
	hitRemoteClientBundle *clients.RemoteClientBundle
	isHit                 bool
}

func (d *Dispatcher) Exchange(query *dns.Msg, inboundIP string) *dns.Msg {
	ClientsBundle := make(map[string]*clients.RemoteClientBundle)
	for k, v := range d.DNSBunch {
		name := k
		ClientsBundle[name] = clients.NewClientBundle(query, v, inboundIP, d.MinimumTTL, d.Cache, name, d.DomainTTLMap)
	}

	var ActiveClientBundle *clients.RemoteClientBundle = nil

	// Priority: client(host), cache, onlyprimary, domain, ip,
	// local hosts, ip
	localClient := clients.NewLocalClient(query, d.Hosts, d.MinimumTTL, d.DomainTTLMap)
	resp := localClient.Exchange()
	if resp != nil {
		// find item in local host/ip list
		return resp
	}
	// local cache
	for _, cb := range ClientsBundle {
		resp := cb.ExchangeFromCache()
		if resp != nil {
			return resp
		}
	}
	// local Domain, ip
	ch := make(chan *HitTask, len(ClientsBundle))
	// hitList := list.New()
	bundleLenght := len(ClientsBundle)
	var c = new(HitTask)
	for bunchName := range ClientsBundle {
		go func(ch chan *HitTask, bunchName string) {
			c.isHit = d.isSelectDomain(ClientsBundle[bunchName], d.DNSFilter[bunchName].DomainList)
			// if c.isHit {
			c.hitRemoteClientBundle = ClientsBundle[bunchName]
			ch <- c
			// }
			return
		}(ch, bunchName)
	}
	for bundleLenght > 0 {
		if x, ok := <-ch; ok {
			log.Info(x.isHit)
			bundleLenght--
			if x.isHit {
				// hitList.PushBack(x.hitRemoteClientBundle)
				ActiveClientBundle = x.hitRemoteClientBundle
				break
				// histList is success hit list, sort by match time
				// use first hit bundle.
			}
		}
	}
	if ActiveClientBundle == nil && d.DefaultDNSBundle == "" {
		log.Info("Domain match failed return nil; DefaultDNSBundle is nil, return nil")
		return nil
	} else if d.DefaultDNSBundle != "" {
		log.Infof("Use default dns bundle: %s", d.DefaultDNSBundle)
		ActiveClientBundle = ClientsBundle[d.DefaultDNSBundle]
	}
	return ActiveClientBundle.Exchange(true, true)
	// isCache bool, isLog bool

	// ActiveClientBundle = d.selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle)

	// if ActiveClientBundle == AlternativeClientBundle {
	// 	resp = ActiveClientBundle.Exchange(true, true)
	// 	// isCache bool, isLog bool
	// 	return resp
	// } else {
	// 	// Only try to Cache result before return
	// 	ActiveClientBundle.CacheResultIfNeeded()
	// 	return ActiveClientBundle.GetResponseMessage()
	// }
}

func (d *Dispatcher) isExchangeForIPv6(query *dns.Msg) bool {
	if query.Question[0].Qtype == dns.TypeAAAA && d.RedirectIPv6Record {
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) isSelectDomain(rcb *clients.RemoteClientBundle, dt matcher.Matcher) bool {
	if dt != nil {
		qn := rcb.GetFirstQuestionDomain()

		if dt.Has(qn) {
			// Find elem in local domain file.
			log.WithFields(log.Fields{
				"DNS":      rcb.Name,
				"question": qn,
				"domain":   qn,
			}).Debug("Matched")
			log.Debugf("Finally use %s DNS", rcb.Name)
			return true
		}

		log.Debugf("Domain %s match fail", rcb.Name)
	} else {
		log.Debug("Domain matcher is nil, not checking")
	}

	return false
}

// func (d *Dispatcher) selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle *clients.RemoteClientBundle) *clients.RemoteClientBundle {

// 	primaryResponse := PrimaryClientBundle.Exchange(false, true)
// 	// isCache bool, isLog bool

// 	if primaryResponse == nil {
// 		log.Debug("Primary DNS return nil, finally use alternative DNS")
// 		return AlternativeClientBundle
// 	}

// 	if primaryResponse.Answer == nil {
// 		if d.WhenPrimaryDNSAnswerNoneUse == "AlternativeDNS" {
// 			log.Debug("Because `WhenPrimaryDNSAnswerNoneUse` configuration, primary DNS response has no answer section but exist, finally use AlternativeDNS")
// 			return AlternativeClientBundle
// 		} else {
// 			log.Debug("Because `WhenPrimaryDNSAnswerNoneUse` configuration, primary DNS response has no answer section but exist, finally use PrimaryDNS")
// 			return PrimaryClientBundle
// 		}
// 	}

// 	for _, a := range PrimaryClientBundle.GetResponseMessage().Answer {
// 		log.Debug("Try to match response ip address with IP network")
// 		var ip net.IP
// 		if a.Header().Rrtype == dns.TypeA {
// 			ip = net.ParseIP(a.(*dns.A).A.String())
// 		} else if a.Header().Rrtype == dns.TypeAAAA {
// 			ip = net.ParseIP(a.(*dns.AAAA).AAAA.String())
// 		} else {
// 			continue
// 		}
// 		if common.IsIPMatchList(ip, d.IPNetworkPrimaryList, true, "primary") {
// 			log.Debug("Finally use primary DNS")
// 			return PrimaryClientBundle
// 		}
// 		if common.IsIPMatchList(ip, d.IPNetworkAlternativeList, true, "alternative") {
// 			log.Debug("Finally use alternative DNS")
// 			return AlternativeClientBundle
// 		}
// 	}
// 	log.Debug("IP network match failed, finally use alternative DNS")
// 	return AlternativeClientBundle
// }
