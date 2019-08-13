// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.
// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package cache implements dns cache feature with edns-client-subnet support.
package cache

// Cache that holds RRs.
// TODO LRU cache link

import (
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// Elem hold an answer and additional section that returned from the cache.
// The signature is put in answer, extra is empty there. This wastes some memory.

type FastMap struct {
	dnsBundle string
	domain    string
}

type elem struct {
	// time added + TTL, after this the elem is invalid
	expiration   time.Time
	msg          *dns.Msg
	fastMap      *FastMap
	next, before *elem
}

// Cache is a cache that holds on the a number of RRs or DNS messages. The cache
// eviction is randomized.
type Cache struct {
	sync.RWMutex

	capacity int
	domain   map[string]*elem
	head     *elem
}

// New returns a new cache with the capacity and the ttl specified.
func New(capacity int) *Cache {
	if capacity <= 0 {
		return nil
	}
	c := new(Cache)
	c.domain = make(map[string]*elem)
	c.capacity = capacity
	c.head = new(elem)
	firstNode := new(elem)
	firstNode.before, firstNode.next = c.head, nil

	c.head.next, c.head.before = firstNode, nil
	return c
}

func (c *Cache) Capacity() int { return c.capacity }

func (c *Cache) Remove(key string) {
	c.Lock()
	delElem := c.domain[key]
	delElem.before.next, delElem.next.before = delElem.next, delElem.before

	delElem.next, delElem.before = nil, nil
	delete(c.domain, key)
	// Use built-in functions for map
	c.Unlock()
}

// EvictRandom removes a random member a the cache.
// Must be called under a write lock.
func (c *Cache) EvictRandom() {
	cacheLength := len(c.domain)
	if cacheLength <= c.capacity {
		return
	}
	i := c.capacity - cacheLength
	for k := range c.domain {
		delete(c.domain, k)
		i--
		if i == 0 {
			break
		}
	}
}

// Insert inserts a DNS message to the Cache. We will cache it for ttl seconds, which
// should be a small (60...300) integer.
func (c *Cache) Insert(key string, m *dns.Msg, mTTL uint32, dnsBundleName string, domainName string) {
	if c.capacity <= 0 || m == nil {
		return
	}
	c.Lock()
	var ttl uint32
	if len(m.Answer) == 0 {
		ttl = mTTL
	} else {
		ttl = m.Answer[0].Header().Ttl
	}
	ttlDuration := time.Duration(ttl) * time.Second
	if _, ok := c.domain[key]; !ok {
		// Insert elem to cache when Cache not have the elem.
		fastMap := new(FastMap)
		fastMap.dnsBundle, fastMap.domain = dnsBundleName, domainName

		newElem := new(elem)
		newElem.expiration, newElem.msg, newElem.fastMap = time.Now().UTC().Add(ttlDuration), m.Copy(), fastMap

		c.head.next.before = newElem

		newElem.next, newElem.before = c.head.next, c.head
		c.head.next = newElem
		c.domain[key] = newElem
	}
	log.Debugf("Cached: %s", key)
	c.Unlock()
}

// Search returns a dns.Msg, the expiration time and a boolean indicating if we found something
// in the cache.
// todo: use finder implementation
func (c *Cache) Search(key string) (*elem, time.Time, bool) {
	if c.capacity <= 0 {
		return nil, time.Time{}, false
	}
	c.RLock()
	if e, ok := c.domain[key]; ok {
		// find elem in cache
		c.RUnlock()

		return e, time.Time{}, true
	}
	c.RUnlock()
	return nil, time.Time{}, false
}

// Key creates a hash key from a question section.
func Key(q dns.Question, ednsIP string) string {
	return fmt.Sprintf("%s %d %s", q.Name, q.Qtype, ednsIP)
}

// Hit returns a dns message from the cache. If the message's TTL is expired nil
// is returned and the message is removed from the cache.
func (c *Cache) Hit(key string, msgid uint16) (isHit bool, BundleName string, _ *dns.Msg) {
	pointer, exp, hit := c.Search(key)
	if hit {
		// Cache hit! \o/
		if time.Since(exp) > 0 {
			pointer.msg.Id = msgid
			pointer.msg.Compress = true
			// Even if something ended up with the TC bit *in* the cache, set it to off
			pointer.msg.Truncated = false
			for _, a := range pointer.msg.Answer {
				a.Header().Ttl = uint32(time.Since(exp).Seconds() * -1)
			}
			return true, pointer.fastMap.dnsBundle, pointer.msg
		}
		// TODO Update program
		return true, pointer.fastMap.dnsBundle, nil
		// Expired! /o\
	}
	return false, "", nil
}

// Dump returns all dns cache information, for dubugging
func (c *Cache) Dump(nobody bool) (rs map[string][]string, cacheLength int) {
	if c.capacity <= 0 {
		return
	}

	cacheLength = len(c.domain)

	rs = make(map[string][]string)

	if nobody {
		return
	}

	c.RLock()
	defer c.RUnlock()

	for k, e := range c.domain {
		var vs []string

		for _, a := range e.msg.Answer {
			vs = append(vs, a.String())
		}
		rs[k] = vs
	}
	return
}
