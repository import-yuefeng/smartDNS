package common

import (
	"net"

	"github.com/import-yuefeng/smartDNS/core/matcher"
)

type Filter struct {
	Matcher       string
	DomainFile    string
	IPNetworkFile string
	IPNetworkList []*net.IPNet
	DomainList    matcher.Matcher
}
