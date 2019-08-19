/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package clients

import (
	"github.com/miekg/dns"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/common"
)

type RemoteClientBundle struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	clients []*RemoteClient

	dnsUpstreams []*common.DNSUpstream
	inboundIP    string
	minimumTTL   int
	domainTTLMap map[string]uint32

	cache *cache.Cache
	Name  string
}

type CacheMessage struct {
	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	MinimumTTL int
	BundleName string
	DomainName string
}

func NewClientBundle(q *dns.Msg, ul []*common.DNSUpstream, ip string, minimumTTL int, cache *cache.Cache, name string, domainTTLMap map[string]uint32) *RemoteClientBundle {

	cb := &RemoteClientBundle{questionMessage: q.Copy(), dnsUpstreams: ul, inboundIP: ip, minimumTTL: minimumTTL, cache: cache, Name: name, domainTTLMap: domainTTLMap}

	for _, u := range ul {

		c := NewClient(cb.questionMessage, u, cb.inboundIP, cb.cache)
		cb.clients = append(cb.clients, c)
	}

	return cb
}

func (cb *RemoteClientBundle) Exchange(isLog bool) *CacheMessage {
	ch := make(chan *RemoteClient, len(cb.clients))
	for _, o := range cb.clients {
		go func(c *RemoteClient, ch chan *RemoteClient) {
			c.Exchange(isLog)
			ch <- c
		}(o, ch)
	}

	var (
		ec *RemoteClient
	)
	cacheMessage := new(CacheMessage)
	for i := 0; i < len(cb.clients); i++ {
		c := <-ch
		if c != nil {
			ec = c
			break
			// use dns that first response
		}
	}
	// defer close(ch)
	if ec != nil && ec.responseMessage != nil {
		cacheMessage.ResponseMessage = ec.responseMessage
		cacheMessage.QuestionMessage = ec.questionMessage

		common.SetMinimumTTL(cacheMessage.ResponseMessage, uint32(cacheMessage.MinimumTTL))
		common.SetTTLByMap(cacheMessage.ResponseMessage, cb.domainTTLMap)
	}

	return cacheMessage
}

func (cb *RemoteClientBundle) IsType(t uint16) bool {
	return t == cb.questionMessage.Question[0].Qtype
}

func (cb *RemoteClientBundle) GetFirstQuestionDomain() string {
	return cb.questionMessage.Question[0].Name[:len(cb.questionMessage.Question[0].Name)-1]
}

func (cb *RemoteClientBundle) GetResponseMessage() *dns.Msg {
	return cb.responseMessage
}
