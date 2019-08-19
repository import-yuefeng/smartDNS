// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cron

import (
	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/detector/ping"
	"github.com/miekg/dns"

	"github.com/robfig/cron"
)

type CacheManager struct {
	cache    *cache.Cache
	interval string
	detector string
}

func NewCacheManager(cache *cache.Cache, interval string, detector string) *CacheManager {
	return &CacheManager{cache, interval, detector}
}

func (cacheManager *CacheManager) Update() {
	head := cacheManager.cache.Head()
	var updateElemNum int
	if cacheManager.cache.Capacity() > cacheManager.cache.Size() {
		updateElemNum = 5
	} else {
		updateElemNum = 5 + cacheManager.cache.Size() - cacheManager.cache.Capacity()
	}

}

func (cacheManager *CacheManager) SwitchDetector(msg *dns.Msg, fastMap *cache.FastMap, DNSBunch map[string][]*common.DNSUpstream) {
	switch cacheManager.detector {
	case "ping":
		return ping.NewPinger(msg, fastMap, DNSBunch)
	default:
		log.Warnf("Detector %s does not exist, using ping as default", cacheManager.detector)
		return ping.NewPinger(msg, fastMap, DNSBunch)
	}
}

func (cacheManager *CacheManager) Crontab() {
	// i := 0
	c := cron.New()
	spec := cacheManager.interval

	c.AddFunc(spec, cacheManager.Update)
	c.Start()

	select {}

}
