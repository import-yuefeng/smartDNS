package cron

import (
	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/miekg/dns"
)

type Task struct {
	msg *dns.Msg
}

type CacheManager struct {
	TaskChan chan *Task
	Cache    *cache.Cache
	Interval string
	Detector string
}
