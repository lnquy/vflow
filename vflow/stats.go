//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    stats.go
//: details: exposes flow status
//: author:  Mehrdad Arshad Rad
//: date:    02/01/2017
//:
//: Licensed under the Apache License, Version 2.0 (the "License");
//: you may not use this file except in compliance with the License.
//: You may obtain a copy of the License at
//:
//:     http://www.apache.org/licenses/LICENSE-2.0
//:
//: Unless required by applicable law or agreed to in writing, software
//: distributed under the License is distributed on an "AS IS" BASIS,
//: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//: See the License for the specific language governing permissions and
//: limitations under the License.
//: ----------------------------------------------------------------------------

package main

import (
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"time"
)

// Common interface for flows statistics.
// All protocols statistics must implement and return this interface in status() method.
type flowStats interface {
	// Dummy methods. Implementers just have to describe this func as its method signature.
	isFlowStats()
}

var startTime = time.Now().Unix()

// StatsSysHandler handles /sys endpoint
func StatsSysHandler(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)
	var data = &struct {
		MemAlloc        uint64
		MemTotalAlloc   uint64
		MemHeapAlloc    uint64
		MemHeapSys      uint64
		MemHeapReleased uint64
		MCacheInuse     uint64
		GCSys           uint64
		GCNext          uint64
		GCLast          string
		NumLogicalCPU   int
		NumGoroutine    int
		MaxProcs        int
		GoVersion       string
		StartTime       string
	}{
		mem.Alloc,
		mem.TotalAlloc,
		mem.HeapAlloc,
		mem.HeapSys,
		mem.HeapReleased,
		mem.MCacheInuse,
		mem.GCSys,
		mem.NextGC,
		time.Unix(0, int64(mem.LastGC)).String(),
		runtime.NumCPU(),
		runtime.NumGoroutine(),
		runtime.GOMAXPROCS(-1),
		runtime.Version(),
		time.Unix(startTime, 0).String(),
	}

	j, err := json.Marshal(data)
	if err != nil {
		logger.Println(err)
	}

	if _, err = w.Write(j); err != nil {
		logger.Println(err)
	}
}

// StatsFlowHandler handles /flow endpoint
func StatsFlowHandler(protos ...proto) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data = &struct {
			StartTime int64
			Stats     map[string]flowStats
		}{
			StartTime: startTime,
			Stats:     make(map[string]flowStats),
		}

		for _, p := range protos {
			data.Stats[p.name()] = p.status()
		}
		j, err := json.Marshal(data)
		if err != nil {
			logger.Println(err)
		}

		if _, err = w.Write(j); err != nil {
			logger.Println(err)
		}
	}
}

func statsHTTPServer(protos ...proto) {
	if !opts.StatsEnabled {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sys", StatsSysHandler)
	mux.HandleFunc("/flow", StatsFlowHandler(protos...))

	addr := net.JoinHostPort(opts.StatsHTTPAddr, opts.StatsHTTPPort)

	logger.Println("starting stats web server ...")
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
