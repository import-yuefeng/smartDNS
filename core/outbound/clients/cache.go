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
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/import-yuefeng/smartDNS/core/cache"
)

type CacheClient struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	ednsClientSubnetIP string

	cache *cache.Cache
}

func NewCacheClient(q *dns.Msg, cache *cache.Cache) *CacheClient {
	// return &CacheClient{questionMessage: q.Copy(), ednsClientSubnetIP: ip, cache: cache}
	return &CacheClient{questionMessage: q.Copy(), cache: cache}

}

func (c *CacheClient) Exchange() (isHit bool, BundleName string, _ *dns.Msg) {
	if c.cache == nil {
		return false, "", nil
	}
	key := cache.Key(c.questionMessage.Question[0], c.ednsClientSubnetIP)
	isHit, bundleName, msg := c.cache.Hit(key, c.questionMessage.Id)
	if isHit {
		log.Debugf("Cache hit: %s", key)
	}
	return isHit, bundleName, msg
}
