package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/doubleforward"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	forwardProxyHost := flag.String("forward-proxy-host", "", "use forward proxy")
	forwardProxyUsername := flag.String("forward-proxy-username", "", "username for forward proxy")
	forwardProxyPassword := flag.String("forward-proxy-password", "", "password for forward proxy")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	if *forwardProxyHost != `` {
		connectDial, err := doubleforward.ConnectDialViaProxy(*forwardProxyHost, *forwardProxyUsername, *forwardProxyPassword)
		if err != nil {
			panic(err)
		}
		proxy.ConnectDial = connectDial
	}

	log.Fatal(http.ListenAndServe(*addr, proxy))
}
