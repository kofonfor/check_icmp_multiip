package main

import (
	"flag"
	"fmt"
	dns "github.com/Focinfi/go-dns-resolver"
	fastping "github.com/tatsushid/go-fastping"
	//"log"
	"net"
	"os"
	"time"
)

type CommandLineConfig struct {
	host_name *string
}

func (*CommandLineConfig) Parse() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

var commandLineCfg = CommandLineConfig{
	host_name: flag.String("host_name", "localhost", "An FQDN to check"),
}

func main() {
	commandLineCfg.Parse()
	dns.Config.SetTimeout(uint(2))
	dns.Config.RetryTimes = uint(4)
	pinger := fastping.NewPinger()
	all_ips := make([]string, 0)
	alive_ips := make(map[string]int)
	dead_ips := make([]string, 0)

	if results, err := dns.Exchange(*commandLineCfg.host_name, "8.8.8.8:53", dns.TypeA); err == nil {
		for _, r := range results {
			all_ips = append(all_ips, r.Content)
			ipaddr, err := net.ResolveIPAddr("ip4:icmp", r.Content)
			if err != nil {
				fmt.Printf("CRITICAL - %v\n", err)
				os.Exit(2)
			}
			pinger.AddIPAddr(ipaddr)
		}
	} else {
		fmt.Printf("CRITICAL - %v\n", err)
		os.Exit(2)
	}
	pinger.MaxRTT = 2000000000 // 2s
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		alive_ips[addr.String()] = 1
	}
	pinger.OnIdle = func() {
	}
	err := pinger.Run()
	if err != nil {
		fmt.Printf("CRITICAL - %v\n", err)
		os.Exit(2)
	}
	for _, ip := range all_ips {
		_, ok := alive_ips[ip]
		if !ok {
			dead_ips = append(dead_ips, ip)
		}
	}
	if len(dead_ips) != 0 {
		fmt.Printf("CRITICAL - some pings (%v/%v) failed\n", len(dead_ips), len(all_ips))
		os.Exit(2)
	} else {
		fmt.Printf("OK - all pings succeeded\n")
		os.Exit(0)
	}
}
