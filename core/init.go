// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.
// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package core implements the essential features.
package core

import (
	"github.com/import-yuefeng/smartDNS/core/config"
	"github.com/import-yuefeng/smartDNS/core/cron"
	"github.com/import-yuefeng/smartDNS/core/inbound"
	"github.com/import-yuefeng/smartDNS/core/outbound"
)

// InitServer func Initiate the server with config file
func InitServer(configFilePath string, smart *bool) {
	conf := config.NewConfig(configFilePath)
	//New dispatcher without RemoteClientBundle, RemoteClientBundle must be initiated when server is running
	dispatcher := outbound.Dispatcher{
		DefaultDNSBundle:   conf.DefaultDNSBundle,
		DNSFilter:          conf.DNSFilter,
		DNSBunch:           conf.DNSBunch,
		RedirectIPv6Record: conf.IPv6UseAlternativeDNS,
		MinimumTTL:         conf.MinimumTTL,
		DomainTTLMap:       conf.DomainTTLMap,
		Hosts:              conf.Hosts,
		Cache:              conf.Cache,
		CacheTimer:         new(cron.CacheUpdate),
		SmartDNS:           *smart,
	}
	s := inbound.NewServer(conf.BindAddress, conf.DebugHTTPAddress, dispatcher, conf.RejectQType)
	if *smart {
		dispatcher.CacheTimer.TaskChan = make(chan *cron.Task, 1000)
		dispatcher.CacheTimer.Cache = dispatcher.Cache
		go dispatcher.CacheTimer.Handle()
	}

	s.Run()

}
