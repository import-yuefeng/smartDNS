package cron

import (
	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/miekg/dns"
)

type Task struct {
	msg *dns.Msg
}

type CacheManager struct {
	TaskChan chan bool
	Cache    *cache.Cache
	Interval string
	TaskSum  int
}
