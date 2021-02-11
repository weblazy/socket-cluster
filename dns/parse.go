package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

// parseTarget takes the user input target string and default port, returns formatted host and port info.
// If target doesn't specify a port, set the port to be the defaultPort.
// If target is in IPv6 format and host-name is enclosed in square brackets, brackets
// are stripped when setting the host.
// examples:
// target: "www.google.com" defaultPort: "443" returns host: "www.google.com", port: "443"
// target: "ipv4-host:80" defaultPort: "443" returns host: "ipv4-host", port: "80"
// target: "[ipv6-host]" defaultPort: "443" returns host: "ipv6-host", port: "443"
// target: ":80" defaultPort: "443" returns host: "localhost", port: "80"
func ParsePort(target, defaultPort string) (host, port string, err error) {
	if target == "" {
		return "", "", errors.New("")
	}
	if ip := net.ParseIP(target); ip != nil {
		// target is an IPv4 or IPv6(without brackets) address
		return target, defaultPort, nil
	}
	if host, port, err = net.SplitHostPort(target); err == nil {
		if port == "" {
			// If the port field is empty (target ends with colon), e.g. "[::1]:", this is an error.
			return "", "", errors.New("")
		}
		// target has port, i.e ipv4-host:port, [ipv6-host]:port, host-name:port
		if host == "" {
			// Keep consistent with net.Dial(): If the host is empty, as in ":80", the local system is assumed.
			host = "localhost"
		}
		return host, port, nil
	}
	if host, port, err = net.SplitHostPort(target + ":" + defaultPort); err == nil {
		// target doesn't have port
		return host, port, nil
	}
	return "", "", fmt.Errorf("invalid target address %v, error info: %v", target, err)
}

// split2 returns the values from strings.SplitN(s, sep, 2).
// If sep is not found, it returns ("", "", false) instead.
func split2(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return "", "", false
	}
	return spl[0], spl[1], true
}

// ParseTarget splits target into a resolver.Target struct containing scheme,
// authority and endpoint.
//
// If target is not a valid scheme://authority/endpoint, it returns {Endpoint:
// target}.
func ParseScheme(target string) (ret Target) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = split2(target, "://")
	if !ok {
		return Target{Endpoint: target}
	}
	ret.Authority, ret.Endpoint, ok = split2(ret.Endpoint, "/")
	if !ok {
		return Target{Endpoint: target}
	}
	return ret
}

func DnsParse(domain string) ([]string, error) {
	target := ParseScheme(domain)
	host, port, _ := ParsePort(target.Endpoint, "80")
	// builder := resolver.Get(target.Scheme)
	// addrs, err := builder.Parse(target)
	fmt.Printf("%#v\n", host)
	fmt.Printf("%#v\n", port)
	return net.DefaultResolver.LookupHost(context.Background(), host)
}
