package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/elazarl/goproxy"
)

type ProxyStatus int

const (
	STOPED ProxyStatus = iota
	Launching
	Running
)

func main() {
	proxyStatusChan := make(chan ProxyStatus)
	listener, err := net.Listen("tcp6", ":0")
	if err != nil {
		panic(err)
	}
	go runProxy(proxyStatusChan, listener)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case nextProxyStatus := <-proxyStatusChan:
			switch nextProxyStatus {
			case Running:
				ipv6Addresses, err := getLocalIpv6Addresses()
				if err != nil {
					log.Fatal(err)
				}
				port := listener.Addr().(*net.TCPAddr).Port
				log.Println("Proxy Server Start!")
				for _, addr := range ipv6Addresses {
					log.Printf("http://[%s]:%d\n", addr, port)
				}
			case Launching:
				log.Println("Proxy Server Launching......")
			case STOPED:
				log.Println("Proxy Server Stoped!")
				listener.Close()
				return
			}
		case sig := <-quit:
			log.Printf("shutting down ... Reason %s \n", sig)
			listener.Close()
			return
		}
	}

}

func runProxy(proxyStatus chan ProxyStatus, listener net.Listener) {
	defer func() {
		proxyStatus <- STOPED
	}()
	proxyStatus <- Launching
	proxy := goproxy.NewProxyHttpServer()
	proxyStatus <- Running
	http.Serve(listener, proxy)
}

func getLocalIpv6Addresses() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ipv6Addresses := []string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err == nil {
			for _, a := range addrs {
				switch v := a.(type) {
				case *net.IPNet:
					if v.IP.IsGlobalUnicast() && v.IP.To4() == nil {
						ipv6Addresses = append(ipv6Addresses, v.IP.String())
					}
				}
			}
		}
	}
	return ipv6Addresses, nil
}
