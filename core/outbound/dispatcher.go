// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package outbound

import (
	"net"

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

type BundleMsg struct {
	result     *clients.CacheMessage
	bundleName string
}

type Bundle struct {
	ClientBundle map[string]*clients.RemoteClientBundle
}

type HitTask struct {
	hitRemoteClientBundle *clients.RemoteClientBundle
	isHit                 bool
}

func (d *Dispatcher) Exchange(query *dns.Msg, inboundIP string) *dns.Msg {
	bundle := new(Bundle)
	bundle.ClientBundle = make(map[string]*clients.RemoteClientBundle)
	for name, v := range d.DNSBunch {
		bundle.ClientBundle[name] = clients.NewClientBundle(query, v, inboundIP, d.MinimumTTL, d.Cache, name, d.DomainTTLMap)
	}

	var ActiveClientBundle *clients.RemoteClientBundle
	// Priority: client(host), cache, onlyprimary, domain, ip,
	// local hosts, ip
	localClient := clients.NewLocalClient(query, d.Hosts, d.MinimumTTL, d.DomainTTLMap)
	resp := localClient.Exchange()
	if resp != nil {
		// find item in local host/ip list
		return resp
	}

	// Global cache(be shared all DNSBunch)
	cacheClient := clients.NewCacheClient(query, d.Cache)
	isHit, bundleName, msg := cacheClient.Exchange()
	if isHit {
		if msg != nil {
			return msg
		} else if msg == nil && bundleName != "" {
			log.Infof("Hit Cache, msg is expiration, but bundleName: %s\n", bundleName)
			ActiveClientBundle = bundle.ClientBundle[bundleName]
			if result := ActiveClientBundle.Exchange(true); result != nil {
				result.BundleName = ActiveClientBundle.Name
				d.CacheResultIfNeeded(result)
				return result.ResponseMessage
			}
		}
	}

	// local Domain, ip
	ch := make(chan *HitTask, len(bundle.ClientBundle))
	bundleLenght := len(bundle.ClientBundle)
	var c = new(HitTask)

	for bunchName := range bundle.ClientBundle {
		go func(ch chan *HitTask, bunchName string) {
			c.isHit = d.isSelectDomain(bundle.ClientBundle[bunchName], d.DNSFilter[bunchName].DomainList)
			c.hitRemoteClientBundle = bundle.ClientBundle[bunchName]
			ch <- c
			return
		}(ch, bunchName)
	}

	for bundleLenght > 0 {
		if x, ok := <-ch; ok {
			bundleLenght--
			if x.isHit {
				ActiveClientBundle = x.hitRemoteClientBundle
				close(ch)
				break
			}
		}
	}
	if ActiveClientBundle == nil {
		log.Warnf("Domain match failed. will check ip list or use default DNS: %s(If not nil)", d.DefaultDNSBundle)
		if resp := d.selectByIPNetwork(bundle); resp != nil {
			log.Info("Match ip!")
			d.CacheResultIfNeeded(resp.result)
			return resp.result.ResponseMessage
		}
	}
	if ActiveClientBundle == nil && d.DefaultDNSBundle != "" {
		log.Warnf("Use default dns bundle: %s", d.DefaultDNSBundle)
		ActiveClientBundle = bundle.ClientBundle[d.DefaultDNSBundle]
	}
	if result := ActiveClientBundle.Exchange(true); result != nil {
		result.BundleName = ActiveClientBundle.Name
		d.CacheResultIfNeeded(result)
		return result.ResponseMessage
	}

	return nil

}

func (d *Dispatcher) CacheResultIfNeeded(cacheMessage *clients.CacheMessage) {
	if d.Cache != nil && cacheMessage.ResponseMessage != nil {
		key := cache.Key(cacheMessage.QuestionMessage.Question[0])
		d.Cache.Insert(key, cacheMessage.ResponseMessage, uint32(cacheMessage.MinimumTTL), cacheMessage.BundleName, cacheMessage.DomainName)
	}
	return
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

func (d *Dispatcher) selectByIPNetwork(bundle *Bundle) *BundleMsg {

	ch := make(chan *BundleMsg, len(bundle.ClientBundle))
	Response := make(map[string]*clients.CacheMessage)

	for name, c := range bundle.ClientBundle {
		go func(c *clients.RemoteClientBundle, ch chan *BundleMsg) {
			result := c.Exchange(true)
			ch <- &BundleMsg{result, name}
			return
		}(c, ch)
	}

	for i := 0; i < len(bundle.ClientBundle); {
		if c := <-ch; c != nil {
			Response[c.bundleName] = c.result
			i++
		}
		if i >= int(float64(len(bundle.ClientBundle))*0.75) {
			close(ch)
			break
		}
	}
	for bundleName, a := range Response {
		for i := range a.ResponseMessage.Answer {
			log.Debug("Try to match response ip address with IP network")
			var ip net.IP

			if a.ResponseMessage.Answer[i].Header().Rrtype == dns.TypeA {
				ip = net.ParseIP(a.ResponseMessage.Answer[i].(*dns.A).A.String())
			} else if a.ResponseMessage.Answer[i].Header().Rrtype == dns.TypeAAAA {
				ip = net.ParseIP(a.ResponseMessage.Answer[i].(*dns.AAAA).AAAA.String())
			} else {
				continue
			}
			if common.IsIPMatchList(ip, d.DNSFilter[bundleName].IPNetworkList, true, bundleName) {
				log.Debugf("(IPMatcher)Finally use: %s", bundleName)
				return &BundleMsg{a, bundleName}
			}
		}
		log.Debug("IP network match failed, return nil")
	}
	return nil

}
