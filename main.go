package main

import (
	"container/ring"
	"errors"
	"flag"
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/peterbourgon/ff"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var ProxyListURL string
var ProxyRing *ring.Ring
var UserPassword string

func getSocks5Proxy() ([]string, error) {
	resp, err := http.Get(ProxyListURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("response status code is not 200: %d", resp.StatusCode))
	}
	defer resp.Body.Close()
	txtFile, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	proxiesUnique := make(map[string]bool)
	for _, p := range strings.Split(string(txtFile), "\n") {
		proxiesUnique[strings.TrimSpace(p)] = true
	}
	proxies := make([]string, 0)
	for p := range proxiesUnique {
		if p != "" {
			proxies = append(proxies, p)
		}
	}

	return proxies, nil
}

func loadProxyListEvery(period time.Duration) {
	for {
		ps, err := getSocks5Proxy()
		if err != nil {
			fmt.Println("Proxy list loading error:", err)
		}

		ProxyRing = ring.New(len(ps))
		for _, p := range ps {
			ProxyRing.Value = p
			ProxyRing = ProxyRing.Next()
		}
		fmt.Println("Proxy list loaded. Count: ", len(ps))
		time.Sleep(period)
	}
}

func getProxy() string {
	if ProxyRing.Len() == 0 {
		return ""
	}
	nextProxy := ProxyRing.Value.(string)
	ProxyRing = ProxyRing.Next()
	return nextProxy
}

func getAuth() *proxy.Auth {
	if UserPassword != "" {
		up := strings.Split(UserPassword, ":")
		return &proxy.Auth{User: up[0], Password: up[1]}
	}
	return nil
}

// Dial with socks5 server from proxy list
func dialOverSocks5(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyAddr := getProxy()
	if proxyAddr == "" {
		return nil, errors.New("proxy list is empty")
	}
	dialSocksProxy, err := proxy.SOCKS5("tcp", proxyAddr, getAuth(), proxy.Direct)

	if err != nil {
		return nil, err
	}
	return dialSocksProxy.Dial(network, addr)

}

func runProxyServer(addr string, port int, user string) error {
	var conf *socks5.Config

	if user != "" {
		up := strings.Split(user, ":")
		conf = &socks5.Config{Dial: dialOverSocks5, Credentials: socks5.StaticCredentials{up[0]: up[1]}}
	} else {
		conf = &socks5.Config{Dial: dialOverSocks5}
	}

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	return server.ListenAndServe("tcp", fmt.Sprintf("%s:%d", addr, port))
}

func main() {
	var addr string
	var port int
	var period time.Duration
	var localUserPassword string

	fs := flag.NewFlagSet("2 level proxy server. Update proxy list every n seconds.", flag.ExitOnError)

	fs.StringVar(&ProxyListURL, "url", "", "URL to proxy list. Prior then file.")
	fs.StringVar(&localUserPassword, "u", "", "socks5 server incoming auth user:password")
	fs.StringVar(&UserPassword, "ur", "", "socks5 servers outgoing auth user:password")
	fs.StringVar(&addr, "a", "0.0.0.0", "listening ip")
	fs.IntVar(&port, "p", 8181, "listening port")
	fs.DurationVar(&period, "n", time.Hour*24, "update period in seconds")

	err := ff.Parse(fs, os.Args[1:])
	if err != nil {
		panic(err)
	}

	if ProxyListURL == "" {
		panic("Specify URL to proxy list!")
	}

	go loadProxyListEvery(period)

	err = runProxyServer(addr, port, localUserPassword)
	if err != nil {
		panic(err)
	}
}
