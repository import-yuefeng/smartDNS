// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package ping

import (
	"container/list"
	"reflect"
	"strings"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/outbound/clients"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/sparrc/go-ping"
)

type Pinger struct {
	fastMap *cache.FastMap
	bundle  map[string]*clients.RemoteClientBundle
}

type BundleMsg struct {
	msg        *dns.Msg
	bundleName string
}

func NewDetector(msg *dns.Msg, fastMap *cache.FastMap, bundle map[string]*clients.RemoteClientBundle) *Pinger {
	return &Pinger{fastMap, bundle}
}

func (data *Pinger) Sort(fastTable *list.List) (fastMap *cache.FastMap) {
	defer func() *cache.FastMap {
		if err := recover(); err != nil {
			log.Error(err)
		}
		return nil
	}()

	if reflect.DeepEqual(fastTable, list.New()) {
		return fastTable.Front().Value.(*cache.FastMap)
	}
	return nil
}

func (data *Pinger) Detect() (fastTable *list.List) {
	fastTable = list.New()
	domainToIP := make(map[string]string)
	var ch chan *BundleMsg
	ch = make(chan *BundleMsg, len(data.bundle))

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
		return
	}()

	for name, c := range data.bundle {
		go func(c *clients.RemoteClientBundle, ch chan *BundleMsg, bundleName string) {
			result := c.Exchange(true)
			ch <- &BundleMsg{result.ResponseMessage, bundleName}
			return
		}(c, ch, name)
	}
	for i := 0; i < len(data.bundle); {
		if c := <-ch; c != nil {
			if c.msg == nil && len(c.msg.Answer) == 0 && c.msg.Answer == nil {
				continue
			}
			log.Info(c.msg.Answer)
			log.Info(c.msg.Answer[0])
			dnsRespon := strings.Fields(c.msg.Answer[0].String())
			domainToIP[c.bundleName] = dnsRespon[4]
			i++
		}
		if i >= int(float64(len(data.bundle))*0.6) {
			break
		}
	}

	var bundle chan string
	bundle = make(chan string, len(data.bundle))

	for name, ip := range domainToIP {
		go func(ip string, name string, bundle chan string) {
			pinger, err := ping.NewPinger(ip)
			if err != nil {
				log.Error(err)
				return
			}
			pinger.Count = 1
			pinger.Run()
			stat := pinger.Statistics()
			log.Info(stat)
			bundle <- name
			return
		}(ip, name, bundle)
	}

	for i := 0; i < len(data.bundle); {
		if bundleName := <-bundle; bundleName != "" {
			fastMap := new(cache.FastMap)
			fastMap.Domain, fastMap.DnsBundle = data.fastMap.Domain, bundleName
			fastTable.PushBack(fastMap)
			i++
		}
		if i >= int(float64(len(data.bundle))*0.6) {
			break
		}
	}
	return fastTable
}
