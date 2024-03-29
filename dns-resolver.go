package main

import (
	"github.com/miekg/dns"
	// "log"
	"errors"
	"net"
	"strings"
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

type domainBase struct {
	label     string
	zone      string
	mname     string
	updateble bool
}

func (r DnsResolver) GetDomainBase(q string) (domainBase, error) {

	// This could be more robust, I may change this to following the
	// the labels back until we find the longest set of labels that
	// produces an actual SOA, there seems to be inconsistancy from
	// DNS servers about what they are happy to return, sometimes
	// no authority.
	var target domainBase
	target.updateble = false
	labels := strings.Split(dns.Fqdn(q), ".")
	// log.Printf("Length of labels: %v\n", len(labels))
	// Do SOA query on zone.
	// Setup the question

	for i := 0; i+1 < len(labels); i++ {

		qname := strings.Join(labels[i:], ".")
		// log.Printf("Constructioning SOA query %v\n", qname)
		m1 := new(dns.Msg)
		m1.SetEdns0(1280, true)
		m1.SetQuestion(qname, dns.TypeSOA)
		a1p, _, err := r.c.Exchange(m1, r.config.Servers[0]+":"+r.config.Port)
		if err != nil {
			return target, err
		}
		// log.Printf("Received RCODE of %v", a1p.Rcode)
		// if a1p.Rcode != dns.RcodeSuccess {
		//	return target, nil
		// }

		// Check answer for SOA record
		//   If found set zone name
		// get MNAME
		//   If found set MNAME (create new dns client?)
		for _, k := range a1p.Answer {
			if c, ok := k.(*dns.SOA); ok {
				// log.Println("Received SOA in response.")
				// log.Printf("MNAME %v, Serial %v, Zone %v", c.Ns, c.Serial, c.Hdr.Name)
				target.zone = c.Hdr.Name
				target.mname = c.Ns
				target.label = strings.Join(labels[:i], ".")
			}
		}
		if target.mname != "" {
			break
		}
	}
	// Requery MNAME for hostname
	m2 := new(dns.Msg)
	m2.RecursionDesired = false
	m2.SetEdns0(1280, true)
	m2.SetQuestion(dns.Fqdn(q), dns.TypeANY)

	// log.Printf("Sending ANY query to %v", target.mname)
	a2p, _, err := r.c.Exchange(m2, target.mname+":53")
	//   If fails, bail on updates.
	//    Otherwise set, updatable to true.
	if err != nil {
		// log.Println(err)
		return target, errors.New("Error talking to MNAME " + target.mname)
	}
	if a2p.Rcode != dns.RcodeSuccess {
		return target, errors.New("Error with MNAME " + target.mname + " responded with " + dns.RcodeToString[dns.RcodeSuccess])
	}
	// log.Printf("Listed MNAME responded with %v!\n", dns.RcodeToString[dns.RcodeSuccess])
	target.updateble = true

	target.label = strings.TrimSuffix(dns.Fqdn(q), "."+target.zone)

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
