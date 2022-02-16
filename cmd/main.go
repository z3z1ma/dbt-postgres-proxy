package main

import (
	"flag"
	"log"
	"net"

	proxy "dbt-pg-proxy"
)

var options struct {
	listenAddress string
	upstream      string
	dbtHost       string
	dbtPort  	  int
}

func main() {
	flag.StringVar(&options.listenAddress, "listen", "0.0.0.0:6432", "Listen address")
	flag.StringVar(&options.upstream, "upstream", "127.0.0.1:5432", "Upstream postgres server")
	flag.StringVar(&options.dbtHost, "dbtHost", "127.0.0.1", "dbt rpc server host")
	flag.IntVar(&options.dbtPort, "dbtPort", 8580, "dbt rpc server port")
	flag.Parse()

	var rewriterFactory proxy.QueryRewriterFactory
	var err error
	rewriterFactory = proxy.NewDbtRewriterFactory(options.dbtHost, options.dbtPort)

	// We should perform an rpc health check here, preferably in a periodic goroutine
	// since the services are interdependent

	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s", options.listenAddress)
	defer ln.Close()
	log.Fatal(proxy.RunProxy(ln, options.upstream, rewriterFactory))
}
