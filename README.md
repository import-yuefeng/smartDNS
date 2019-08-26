<p align="right">En | <a href="https://github.com/import-yuefeng/smartDNS/blob/master/README-CN.md">中文简体</a>

# smartDNS (based on overture)

<img src="https://github.com/import-yuefeng/smartDNS/blob/master/smartDNS.png" width="150">

----

[![Build Status](https://travis-ci.com/import-yuefeng/smartDNS.svg)](https://travis-ci.com/import-yuefeng/smartDNS)
[![GoDoc](https://godoc.org/github.com/import-yuefeng/smartDNS?status.svg)](https://godoc.org/github.com/import-yuefeng/smartDNS)
[![Go Report Card](https://goreportcard.com/badge/github.com/import-yuefeng/smartDNS)](https://goreportcard.com/report/github.com/import-yuefeng/smartDNS)

smartDNS is a smart DNS server/forwarder/dispatcher written in Go.

smartDNS based on [overture](https://github.com/shawn1m/overture).


**Please note: If you are using the binary releases, please follow the instructions in the README file with
corresponding git version tag. The README in master branch are subject to change and does not always reflect the correct
 instructions to your binary release version.**

## Features

+ Support DNSbunch(All DNS bundle)
+ Support DNSbundle(A group of similar DNS server)
  example: HK-DNS, CN-DNS, US-DNS
+ Support DNS cache update automatically(FastTable)
+ 
+ Dispatcher
    + Custom domain
    + Custom IP network


### Dispatch process

Overture can force custom domain DNS queries to use selected DNS when applicable.

For custom IP network, overture will query the domain with primary DNS firstly. If the answer is empty or the IP
is not matched then overture will finally use the alternative DNS servers.

## Structure
<img src="https://github.com/import-yuefeng/smartDNS/blob/master/smartDnsStructure.png" width="300">

## Installation

You can download binary releases from the [release](https://github.com/import-yuefeng/smartDNS/releases).


## Usages

Start with the default config file -> ./config.json

    $ ./smartDNS

Or use your own config file:

    $ ./smartDNS -c /path/to/config.json

Verbose mode:

    $ ./smartDNS -v

Log to file:

    $ ./smartDNS -l /path/to/overture.log

For other options, please see help:

    $ ./smartDNS -h

Tips:

+ Root privilege is required if you are listening on port 53.

###  Configuration Syntax

Configuration file is "config.json" by default:

```json
{
  "BindAddress": ":53",
  "DebugHTTPAddress": "127.0.0.1:5555",
  "DNSBunch": {
    "HK-DNS": [
      {
        "Name": "Microsoft-HK",
        "Address": "47.91.128.195:53",
        "Protocol": "udp",
        "SOCKS5Address": "",
        "Timeout": 6,
        "EDNSClientSubnet": {
          "Policy": "disable",
          "ExternalIP": "",
          "NoCookie": false
        }
      },
      {
        "Name": "Google-HK",
        "Address": "8.8.8.8:53",
        "Protocol": "udp",
        "SOCKS5Address": "",
        "Timeout": 3,
        "EDNSClientSubnet": {
          "Policy": "enable",
          "ExternalIP": "",
          "NoCookie": false
        }
      }
    ],
    "CN-DNS": [
      {
        "Name": "ChinaTelecom-CN",
        "Address": "114.114.114.114:53",
        "Protocol": "udp",
        "SOCKS5Address": "",
        "Timeout": 6,
        "EDNSClientSubnet": {
          "Policy": "disable",
          "ExternalIP": "",
          "NoCookie": false
        }
      },
      {
        "Name": "Baidu-CN",
        "Address": "180.76.76.76:53",
        "Protocol": "udp",
        "SOCKS5Address": "",
        "Timeout": 6,
        "EDNSClientSubnet": {
          "Policy": "disable",
          "ExternalIP": "",
          "NoCookie": false
        }
      }
    ]
  },
  "DNSFilter": {
    "HK-DNS": {
      "IPNetworkFile": "",
      "DomainFile": "configuration/hk.domain",
      "Matcher": "regex-list"
    },
    "CN-DNS": {
      "IPNetworkFile": "",
      "DomainFile": "configuration/cn.domain",
      "Matcher": "regex-list"
    }
  },
  "IPv6UseAlternativeDNS": false,
  "HostsFile": "./hosts",
  "MinimumTTL": 0,
  "DomainTTLFile": "./domain_ttl_sample",
  "CacheSize": 0,
  "RejectQType": [
    255
  ]
}```

```

## Acknowledgements
+ Fork:
    + [overture](https://github.com/shawn1m/overture): MIT
+ Dependencies:
    + [dns](https://github.com/miekg/dns): BSD-3-Clause
    + [logrus](https://github.com/Sirupsen/logrus): MIT
+ Code reference:
    + [skydns](https://github.com/skynetservices/skydns): MIT
    + [go-dnsmasq](https://github.com/janeczku/go-dnsmasq):  MIT
+ Contributors: https://github.com/import-yuefeng/smartDNS/graphs/contributors

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
