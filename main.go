package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var cert, key, port, tlsPort string
	flag.StringVar(&cert, "cert", "", "cert file") // certs/cert.pem
	flag.StringVar(&key, "key", "", "key file")    // certs/key.pem
	flag.StringVar(&tlsPort, "tls-port", "8443", "tls port")
	flag.StringVar(&port, "port", "8080", "port")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	s := Server{
		port:     ":" + port,
		tlsPort:  ":" + tlsPort,
		Sessions: map[string]*Session{},
		cert:     cert,
		key:      key,
	}
	s.start()
}
