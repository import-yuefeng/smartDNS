// based on shawn1m/overture (MIT LICENSE)
// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package main is the entry point of whole program.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/import-yuefeng/smartDNS/core"
)

var (
	version = "0.0.1"

	configPath      = flag.String("c", "./config.json", "config file path")
	logPath         = flag.String("l", "", "log file path")
	isLogVerbose    = flag.Bool("v", false, "verbose mode")
	processorNumber = flag.Int("p", runtime.NumCPU(), "number of processor to use")
	isShowVersion   = flag.Bool("V", false, "current version of smartDNS")
	query           = flag.String("q", "", "query DNS")
)

func main() {
	flag.Parse()
	// parse command-line flag
	if *isShowVersion {
		fmt.Println(version)
		// println smartDNS version
		return
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		// except for program has Warn
		log.SetLevel(log.WarnLevel)
	}

	if *logPath != "" {
		lf, err := os.OpenFile(*logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			log.Errorf("Unable to open log file for writing: %s", err)
		} else {
			log.SetOutput(io.MultiWriter(lf, os.Stdout))
		}
	}

	log.Infof("smartDNS %s", version)
	log.Info("If you need any help, please visit the project repository: https://github.com/import-yuefeng/smartDNS")

	runtime.GOMAXPROCS(*processorNumber)

	core.InitServer(*configPath)
}
