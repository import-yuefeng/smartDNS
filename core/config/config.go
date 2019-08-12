// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package config

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/import-yuefeng/smartDNS/core/cache"
	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/hosts"
	"github.com/import-yuefeng/smartDNS/core/matcher"
	"github.com/import-yuefeng/smartDNS/core/matcher/full"
	"github.com/import-yuefeng/smartDNS/core/matcher/mix"
	"github.com/import-yuefeng/smartDNS/core/matcher/regex"
	"github.com/import-yuefeng/smartDNS/core/matcher/suffix"
)

type Config struct {
	BindAddress           string
	DebugHTTPAddress      string
	IPv6UseAlternativeDNS bool
	DefaultDNSBundle      string
	HostsFile             string
	MinimumTTL            int
	DomainTTLFile         string
	CacheSize             int
	RejectQType           []uint16
	DomainTTLMap          map[string]uint32
	Hosts                 *hosts.Hosts
	Cache                 *cache.Cache
	DNSFilter             map[string]*common.Filter
	DNSBunch              map[string][]*common.DNSUpstream
}

// NewConfig will input configFile(json) path, output *Config stuct
func NewConfig(configFile string) *Config {
	// call parseJson func to parse json file
	config := parseJSON(configFile)

	config.DomainTTLMap = getDomainTTLMap(config.DomainTTLFile)
	// configure will load all DNS filter rule
	for k, _ := range config.DNSFilter {
		config.DNSFilter[k].DomainList = initDomainMatcher(config.DNSFilter[k].DomainFile, config.DNSFilter[k].Matcher)
		config.DNSFilter[k].IPNetworkList = getIPNetworkList(config.DNSFilter[k].IPNetworkFile)
	}

	if config.MinimumTTL > 0 {
		// check MinimumTTL value, manual define MinimumTTL
		// MinimumTTL is disabled when MinimumTTL is zero(default)
		log.Infof("Minimum TTL has been set to %d", config.MinimumTTL)
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.Cache = cache.New(config.CacheSize)
	if config.CacheSize > 0 {
		// check CacheSize value, manual define cache capacity
		// CacheSize is disabled when CacheSize is zero(default)
		log.Infof("CacheSize is %d", config.CacheSize)
	} else {
		log.Info("Cache is disabled")
	}

	h, err := hosts.New(config.HostsFile)
	if err != nil {
		log.Warnf("Failed to load hosts file: %s", err)
	} else {
		config.Hosts = h
		log.Info("Hosts file has been loaded successfully")
	}

	return config
}

func parseJSON(path string) *Config {
	// parseJSON Read file(json) convert to *Config
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		os.Exit(1)
	}

	j := new(Config)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err)
		os.Exit(1)
	}

	return j
}

func getDomainTTLMap(file string) map[string]uint32 {
	if file == "" {
		return map[string]uint32{}
	}

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open domain TTL file %s: %s", file, err)
		return nil
	}
	defer f.Close()

	successes := 0
	failures := 0
	var failedLines []string

	dtl := map[string]uint32{}

	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("Failed to read domain TTL file %s: %s", file, err)
			} else {
				log.Debugf("Reading domain TTL file %s reached EOF", file)
			}
			break
		}

		if line != "" {
			words := strings.Fields(line)
			if len(words) > 1 {
				tempInt64, err := strconv.ParseUint(words[1], 10, 32)
				dtl[words[0]] = uint32(tempInt64)
				if err != nil {
					log.WithFields(log.Fields{"domain": words[0], "ttl": words[1]}).Warnf("Invalid TTL for domain %s: %s", words[0], words[1])
					failures++
					failedLines = append(failedLines, line)
				}
				successes++
			} else {
				failedLines = append(failedLines, line)
				failures++
			}
		}
	}

	if len(dtl) > 0 {
		log.Infof("Domain TTL file %s has been loaded with %d records (%d failed)", file, successes, failures)
		if len(failedLines) > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	} else {
		log.Warnf("No element has been loaded from domain TTL file: %s", file)
		if len(failedLines) > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	}

	return dtl
}

func getDomainMatcher(name string) (m matcher.Matcher) {
	switch name {
	case "suffix-tree":
		return suffix.DefaultDomainTree()
	case "full-map":
		return &full.Map{DataMap: make(map[string]struct{}, 100)}
	case "full-list":
		return &full.List{}
	case "regex-list":
		return &regex.List{}
	case "mix-list":
		return &mix.List{}
	default:
		log.Warnf("Matcher %s does not exist, using regex-list matcher as default", name)
		return &regex.List{}
	}
}

func initDomainMatcher(file string, name string) (m matcher.Matcher) {
	m = getDomainMatcher(name)

	if file == "" {
		return
	}

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open domain file %s: %s", file, err)
		return nil
	}
	defer f.Close()

	lines := 0
	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("Failed to read domain file %s: %s", file, err)
			} else {
				log.Debugf("Reading domain file %s reached EOF", file)
				break
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			_ = m.Insert(line)
			lines++
		}
	}

	if lines > 0 {
		log.Infof("Domain file %s has been loaded with %d records (%s)", file, lines, m.Name())
	} else {
		log.Warnf("No element has been loaded from domain file: %s", file)
	}

	return
}

func getIPNetworkList(file string) []*net.IPNet {
	ipNetList := make([]*net.IPNet, 0)

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open IP network file: %s", err)
		return nil
	}
	defer f.Close()

	successes := 0
	failures := 0
	var failedLines []string

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("Failed to read IP network file %s: %s", file, err)
			} else {
				log.Debugf("Reading IP network file %s has reached EOF", file)
			}
			break
		}

		if line != "" {
			_, ipNet, err := net.ParseCIDR(strings.TrimSuffix(line, "\n"))
			if err != nil {
				log.Errorf("Error parsing IP network CIDR %s: %s", line, err)
				failures++
				failedLines = append(failedLines, line)
				continue
			}
			ipNetList = append(ipNetList, ipNet)
			successes++
		}
	}

	if len(ipNetList) > 0 {
		log.Infof("IP network file %s has been loaded with %d records", file, successes)
		if failures > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	} else {
		log.Warnf("No element has been loaded from IP network file: %s", file)
		if failures > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	}

	return ipNetList
}
