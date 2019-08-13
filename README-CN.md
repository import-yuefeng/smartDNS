
# smartDNS (基于 overture)

<img src="https://github.com/import-yuefeng/smartDNS/blob/master/smartDNS.png" width="150">

----

[![Build Status](https://travis-ci.com/import-yuefeng/smartDNS.svg)](https://travis-ci.com/import-yuefeng/smartDNS)
[![GoDoc](https://godoc.org/github.com/import-yuefeng/smartDNS?status.svg)](https://godoc.org/github.com/import-yuefeng/smartDNS)
[![Go Report Card](https://goreportcard.com/badge/github.com/import-yuefeng/smartDNS)](https://goreportcard.com/report/github.com/import-yuefeng/smartDNS)

**由于smartDNS 还在开发的早期，请暂时不要将 smartDNS 用于生产环境！**

smartDNS 是一个用 Go 实现的 智能的DNS 服务/转发/调度器 .

smartDNS 基于 [overture](https://github.com/shawn1m/overture).


**请注意: 本 Readme 仅仅确保 master 分支正确，如若使用版本差距较大的 Binary 版本，请注意检查 Release 说明.**

## 功能和未来展望

+ 支持 DNSbunch(所有 DNSbundle的集合)
+ 支持 DNSbundle(一组相似的 DNS 集合)
  example: HK-DNS, CN-DNS, US-DNS
+ 支持 DNS Cache 主动更新 并根据本地网络情况建立 Domain 和 DNS 的映射关系，实现对域名级细度的 DNS 加速 (FastTable)
+ 支持自定义主动探测方法，实现对不同映射关系的改变，优化本地网络情况
+ 支持 EDNS Client Subnet (ECS) [RFC7871](https://tools.ietf.org/html/rfc7871)
+ 核心调度器
    + 任何一组相似的 DNS 集合（DNSbundle）均可以配置 Domain list
    + 任何一组相似的 DNS 集合（DNSbundle）均可以配置 Custom IP network


### Dispatch process

Overture can force custom domain DNS queries to use selected DNS when applicable.

For custom IP network, overture will query the domain with primary DNS firstly. If the answer is empty or the IP
is not matched then overture will finally use the alternative DNS servers.

## 安装

我们提供跨平台多版本的 二进制程序可下载： [release](https://github.com/import-yuefeng/smartDNS/releases).


## Usages

使用默认配置文件 -> ./config.json

    $ ./smartDNS

或者使用自有的配置文件地址:

    $ ./smartDNS -c /path/to/config.json

详细模式:

    $ ./smartDNS -v

打印日志到文件:

    $ ./smartDNS -l /path/to/overture.log

需要使用其他指令，请使用帮助:

    $ ./smartDNS -h

提示:

+ Root privilege is required if you are listening on port 53.

###  配置文件的语法

默认读取的配置文件名为: config.json :

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

## 感谢
+ Fork:
    + [overture](https://github.com/shawn1m/overture): MIT
+ Dependencies:
    + [dns](https://github.com/miekg/dns): BSD-3-Clause
    + [logrus](https://github.com/Sirupsen/logrus): MIT
+ Code reference:
    + [skydns](https://github.com/skynetservices/skydns): MIT
    + [go-dnsmasq](https://github.com/janeczku/go-dnsmasq):  MIT
+ Contributors: https://github.com/import-yuefeng/smartDNS/graphs/contributors

## 开源协议

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
