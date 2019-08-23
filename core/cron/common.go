package cron

import (
	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/miekg/dns"
)

type CacheManager struct {
	c        *cache.Cache
	interval string
	detector string
}

type Task struct {
	msg *dns.Msg
}

type CacheUpdate struct {
	TaskChan chan *Task
	Cache    *cache.Cache
}
