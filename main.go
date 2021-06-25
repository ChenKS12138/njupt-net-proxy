package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/armon/go-socks5"
	"github.com/elazarl/goproxy"
)

type ProxyStatus int

const (
	STOPED ProxyStatus = iota
	Launching
	Running
)

func main() {
	httpPort := flag.Int("httpPort", -1, "to enable http proxy server")
	socks5Port := flag.Int("socks5Port", -1, "to enable socks5 proxy server")

	flag.Parse()

	if (*httpPort) < 0 && (*socks5Port) < 0 {
		fmt.Printf("Usage:\n  njupt-net-proxy -httpPort 1087\n  njupt-net-proxy -socks5Port 1080\n")
		return
	}

	httpProxyStatusChan := make(chan ProxyStatus)
	socks5ProxyStatusChan := make(chan ProxyStatus)

	if (*httpPort) >= 0 {
		listener, err := net.Listen("tcp6", fmt.Sprintf(":%d", *httpPort))
		if err != nil {
			panic(err)
		}
		go func() {
			go runHttpProxy(httpProxyStatusChan, listener)
			for {
				switch <-httpProxyStatusChan {
				case Running:
					ipv6Addres, err := getLocalIpv6Addresses()
					if err != nil {
						log.Fatal(err)
					}
					port := listener.Addr().(*net.TCPAddr).Port
					log.Println("Http Proxy Server Start!")
					for _, addr := range ipv6Addres {
						log.Printf("http://[%s]:%d\n", addr, port)
					}
				case Launching:
					log.Println("Http Proxy Server Launching....")
				case STOPED:
					listener.Close()
					log.Println("Http Proxy Server Stoped!")
				}
			}
		}()
	}

	if (*socks5Port) >= 0 {
		listener, err := net.Listen("tcp6", fmt.Sprintf(":%d", *socks5Port))
		if err != nil {
			panic(err)
		}
		go func() {
			go runSocks5Proxy(socks5ProxyStatusChan, listener)
			for {
				switch <-socks5ProxyStatusChan {
				case Running:
					ipv6Addres, err := getLocalIpv6Addresses()
					if err != nil {
						log.Fatal(err)
					}
					port := listener.Addr().(*net.TCPAddr).Port
					log.Println("Socks5 Proxy Server Start!")
					for _, addr := range ipv6Addres {
						log.Printf("socks5h://[%s]:%d\n", addr, port)
					}
				case Launching:
					log.Println("Socks5 Proxy Server Launching...")
				case STOPED:
					listener.Close()
					log.Println("SOcks5 Proxy Server Stoped!")
				}
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	sig := <-quit
	log.Printf("Shutting down ... \nReason %s \n", sig)
	if (*httpPort) >= 0 {
		httpProxyStatusChan <- STOPED
	}
	if (*socks5Port) >= 0 {
		socks5ProxyStatusChan <- STOPED
	}
}

func runHttpProxy(proxyStatus chan ProxyStatus, listener net.Listener) {
	defer func() {
		proxyStatus <- STOPED
	}()
	proxyStatus <- Launching
	proxy := goproxy.NewProxyHttpServer()
	proxyStatus <- Running
	http.Serve(listener, proxy)
}

func runSocks5Proxy(proxyStatus chan ProxyStatus, listener net.Listener) {
	defer func() {
		proxyStatus <- STOPED
	}()
	proxyStatus <- Launching
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	proxyStatus <- Running
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
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
