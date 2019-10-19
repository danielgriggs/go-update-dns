package main

import (
	"fmt"
	"net"
	"strings"
)

type DnsHost struct {
	HostName    string
	finalName   string
	targetIPs   []net.IP
	Resolveable bool
	cnameChain  []string
	cnameDepth  uint8
	cnameErr    string
}

func DiscoverHost(host string) (DnsHost, error) {
	var h DnsHost
	h.HostName = host
	h.getRecords()

	return h, nil
}

func (h *DnsHost) IsCname() bool {
	return h.cnameDepth > 0
}

func (h *DnsHost) HasIP(ip_str string) bool {
	ip := net.ParseIP(ip_str)
	for _, b := range h.targetIPs {
		if b.Equal(ip) {
			return true
		}
	}
	return false
}

func (h *DnsHost) getUpdateTarget() (*DnsHost, error) {

	// Do SOA query on zone
	// Check auth section.
	// Check answer for zone name
	//   If found set zone name
	// get MNAME
	//   If found set MNAME (create new dns client?)
	// Requery MNAME for hostname
	//   If fails, bail on updates.
	//    Otherwise set, updatable to true.
	return h, nil
}

func (h *DnsHost) getRecords() (*DnsHost, error) {

	if !strings.HasSuffix(h.HostName, ".") {
		h.HostName = h.HostName + "."
	}
	h.finalName = h.HostName

	h.chaseCnames()

	if len(h.targetIPs) == 0 {
		h.Resolveable = false
	}
	h.Resolveable = true

	return h, nil
}

func (h *DnsHost) chaseCnames() {
	// Move this into dns-resolver?
	target = h.HostName
	var (
		// Defaults
		zero       uint8 = 0
		cnameValid bool  = true
		maxDepth   uint8 = 10
	)
	res := NewResolver()
	chain := []string{target}

	for t, d := target, zero; t != "" && cnameValid; d++ {
		// fmt.Printf("start lookup: %v, final: %v\n", t, h.finalName)
		nt, err := res.GetCname(t)
		if err != nil {
			fmt.Println(err)
			cnameValid = false

		}
		cnameValid = true
		if nt != "" {
			for _, d := range chain {
				if d == nt {
					h.cnameErr = fmt.Sprintf("CNAME loop at %v", nt)
				}
			}
			chain = append(chain, nt)
			h.finalName = nt
		}
		if d >= maxDepth {
			h.cnameErr = fmt.Sprintf("CNAME max depth %v exceeded at %v", d, nt)
			cnameValid = false

		}
		t = nt
		h.cnameDepth = d
	}
	h.cnameChain = chain
	// Part Two is take the final target name and resolve that to an IP.
	// fmt.Printf("start A/AAAA lookup: %v\n", h.finalName)
	var a, err = res.GetIPs(h.finalName)
	if err != nil {
		fmt.Println(err)
		cnameValid = false

	}
	h.targetIPs = a
}

func (h *DnsHost) FinalName() string {
	return h.finalName
}

func (h *DnsHost) TargetIP() string {
	var a []string

	for _, ip := range h.targetIPs {
		a = append(a, ip.String())
	}

	return strings.Join(a, ", ")
}

func (h *DnsHost) Stringfy() string {

	if !h.Resolveable {
		return fmt.Sprintf("%v is not resolvable", h.HostName)
	}
	c := ""
	if h.cnameDepth == 0 {
		c = fmt.Sprintf("")
	} else {
		b := strings.Join(h.cnameChain, " -> ")
		c = " via: " + b
	}
	return fmt.Sprintf("%v, ip %v%v", h.HostName, h.TargetIP(), c)
}
