# socks5-list-proxy
Socks5 2 level proxy written on Go, support updating proxy list.


Help:
```bash
./socks5-list-proxy -h

Usage of 2 level proxy server. Update proxy list every n seconds.:
  -a string
    	listening ip (default "0.0.0.0")
  -n duration
    	update period in seconds (default 24h0m0s)
  -p int
    	listening port (default 8181)
  -u string
    	socks5 server incoming auth user:password
  -ur string
    	socks5 servers outgoing auth user:password
  -url string
    	URL to proxy list. Prior then file.
```

Run:
```bash
go run main -url https://path/to/proxies.txt -u test:test
```


Docker run:
```bash
docker build -t socks5-list-proxy .
docker run --rm socks5-list-proxy -url https://path/to/proxies.txt -u test:test
```