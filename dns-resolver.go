package main

import (
	"github.com/miekg/dns"
	"net"
)

type DnsResolver struct {
	config *dns.ClientConfig
	c      *dns.Client
}

func NewResolver() DnsResolver {
	var r DnsResolver
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return r
	}
	r.config = config
	r.c = new(dns.Client)

	return r
}

func (r DnsResolver) GetCname(q string) (string, error) {

	// Setup the question
	m := new(dns.Msg)
	m.SetEdns0(1280, true)
	m.SetQuestion(q, dns.TypeCNAME)

	ap, _, err := r.c.Exchange(m, r.config.Servers[0]+":"+r.config.Port)
	if err != nil {
		return "", err
	}
	if ap.Rcode != dns.RcodeSuccess {
		return "", nil
	}

	var target string
	for _, k := range ap.Answer {
		if c, ok := k.(*dns.CNAME); ok {
			target = c.Target
		}
	}

	return target, nil
}

func (r DnsResolver) GetIPs(q string) ([]net.IP, error) {

	// Setup the question
	m := new(dns.Msg)
	m.SetEdns0(1280, true)
	m.SetQuestion(q, dns.TypeA)

	// Setup the answer
	a := make([]net.IP, 0)

	ap, _, err := r.c.Exchange(m, r.config.Servers[0]+":"+r.config.Port)
	if err != nil {
		return a, err
	}
	if ap.Rcode != dns.RcodeSuccess {
		return a, nil
	}

	for _, ans := range ap.Answer {
		if arec, ok := ans.(*dns.A); ok {
			ip := arec.A
			a = append(a, ip)
		}
	}

	m.SetQuestion(q, dns.TypeAAAA)

	a4p, _, err := r.c.Exchange(m, r.config.Servers[0]+":"+r.config.Port)
	if err != nil {
		return a, err
	}
	if a4p.Rcode != dns.RcodeSuccess {
		return a, nil
	}

	for _, ans := range a4p.Answer {
		if arec, ok := ans.(*dns.AAAA); ok {
			ip := arec.AAAA
			a = append(a, ip)
		}
	}
	return a, nil
}
