// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cron

import (
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func (cacheManager *CacheManager) AutoUpdate() {

	var updateElemNum, delNum int
	updateElemNum = cacheManager.Cache.Capacity() / 10
	// cron updater will update 10%(capacity) elems
	if cacheManager.Cache.Size() > cacheManager.Cache.Capacity() {
		// lru-list is too long
		delNum = cacheManager.Cache.Size() - cacheManager.Cache.Capacity()
	}
	if delNum != 0 {
		cacheManager.Cache.RemoveTail(delNum)
	}
	log.Infof("Cache info current: %d, updateTask: %d, Capacity: %d.", cacheManager.Cache.Size(), updateElemNum, cacheManager.Cache.Capacity())
	// TODO add update func
	cacheManager.Sort()

}

func (cacheManager *CacheManager) Sort() {
	return
}

func (cacheManager *CacheManager) Crontab() {
	// i := 0
	c := cron.New()
	spec := cacheManager.Interval

	c.AddFunc(spec, cacheManager.AutoUpdate)
	c.Start()

	select {}

}
