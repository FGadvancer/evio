// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tidwall/evio"
)

func main() {
	var port int
	var loops int
	var udp bool
	var trace bool
	var reuseport bool
	var stdlib bool

	flag.IntVar(&port, "port", 50000, "server port")
	flag.BoolVar(&udp, "udp", false, "listen on udp")
	flag.BoolVar(&reuseport, "reuseport", false, "reuseport (SO_REUSEPORT)")
	flag.BoolVar(&trace, "trace", true, "print packets to console")
	flag.IntVar(&loops, "loops", 0, "num loops")
	flag.BoolVar(&stdlib, "stdlib", false, "use stdlib")
	flag.Parse()

	var events evio.Events
	events.NumLoops = loops
	events.Serving = func(srv evio.Server) (action evio.Action) {
		log.Printf("echo server started on port %d (loops: %d)", port, srv.NumLoops)
		if reuseport {
			log.Printf("reuseport")
		}
		if stdlib {
			log.Printf("stdlib")
		}
		return
	}
	goBinFullPath := os.Getenv("HARMONY_GO_BIN")
	if len(goBinFullPath) == 0 {
		goBinFullPath = "go" // Default Go binary path
		log.Println("not find env variable: HARMONY_GO_BIN ,Use Default:", goBinFullPath)
	} else {
		log.Println("find env variable: HARMONY_GO_BIN ,Use:", goBinFullPath)
	}
	events.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		if trace {
			log.Printf("opened %s", c.RemoteAddr())
		}
		if reuseport {
			opts.ReuseInputBuffer = true
		}
		if stdlib {
			c.SetContext("stdlib")
		} else {
			c.SetContext("evio")
		}
		return

	}
	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		if trace {
			log.Printf("receive data from client %s", strings.TrimSpace(string(in)))
		}
		out = in
		return
	}
	scheme := "tcp"
	if udp {
		scheme = "udp"
	}
	if stdlib {
		scheme += "-net"
	}
	log.Fatal(evio.Serve(events, fmt.Sprintf("%s://:%d?reuseport=%t", scheme, port, reuseport)))
}
