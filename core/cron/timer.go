// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cron

import (
	"time"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/detector/ping"
	"github.com/import-yuefeng/smartDNS/core/outbound/clients"
	log "github.com/sirupsen/logrus"
)

func (worker *CacheManager) AddTask(expiration uint32, cacheMessage *clients.CacheMessage, fastMap *cache.FastMap, bundle map[string]*clients.RemoteClientBundle) {
	if expiration < 100 {
		// Set minimum update time
		expiration = 100
	}
	newTimer := time.NewTimer(time.Second * time.Duration(expiration))
	go func() {
		<-newTimer.C
		key := cache.Key(cacheMessage.ResponseMessage.Question[0])
		if _, _, ok := worker.Cache.Search(key); !ok {
			return
		}
		TaskDetector := ping.NewDetector(cacheMessage.ResponseMessage, fastMap, bundle)
		fastMapList := TaskDetector.Detect()
		fastMap := TaskDetector.Sort(fastMapList)

		if fastMap != nil && key != "" {
			if success := worker.Cache.Update(key, fastMap); !success {
				worker.Cache.Insert(key, cacheMessage.ResponseMessage, uint32(cacheMessage.MinimumTTL), cacheMessage.BundleName, cacheMessage.DomainName)
			}
			log.Infof("task: %s , time: %d expired\n", cacheMessage.ResponseMessage.Answer, expiration)
			go worker.AddTask(expiration, cacheMessage, fastMap, bundle)
		} else {
			log.Warn("fastMap is nil!")
		}
	}()

}

func (worker *CacheManager) Handle() {
	log.Info("Start CacheUpdate handle program\n")
	for {
		select {
		case _, ok := <-worker.TaskChan:
			if !ok {
				break
			}
		}
	}

}
