package main

import "socket-cluster/dns"

func main() {
	dns.DnsParse("dns:///:443")
}
