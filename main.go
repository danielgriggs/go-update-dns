package main

import (
	"flag"
	"log"
	"strings"
	// "github.com/urfave/cli"
)

var name string
var ptr bool
var target string

func init() {

	flag.BoolVar(&ptr, "ptr", false, "Create or delete ptr record if possible (bool)")
	// Target flag
	flag.StringVar(&target, "target", "", "Target of the new record, ip address or A record.")
	flag.StringVar(&target, "t", "", "short form of --target")
	// Create flag
	var create string
	flag.StringVar(&create, "create", "", "Only create a new record of this name")
	flag.StringVar(&create, "c", "", "short form of --create")
	// Update flag
	var update string
	flag.StringVar(&update, "update", "", "Update this record if it already exists")
	flag.StringVar(&update, "u", "", "short form of --update")
	// Delete flag
	var delete string
	flag.StringVar(&delete, "delete", "", "Delete this record if it exists or not")
	flag.StringVar(&delete, "d", "", "short form of --delete")
	// Lookup flag
	var lookup string
	flag.StringVar(&lookup, "lookup", "", "Just check if this record exists (target optional)")
	flag.StringVar(&lookup, "l", "", "short form of --lookup")

}

func main() {
	flag.Parse()
	// log.Println("Action:", action, " Record:", name, "Target:", target)
	// Desired state
	var desiredstate = ProcessFlags()

	// First me lookup the provided record to see if it exists already.
	prestate, err := DiscoverHost(desiredstate.Record)
	if err != nil {
		log.Println(err)
	}

	switch desiredstate.Action {
	case "lookup":
		log.Printf("Looking up record to check state")
		log.Println(prestate.Stringfy())
	default:
		log.Printf("Unknown action %v", desiredstate.Action)
		compareState(prestate, desiredstate)
		log.Println(prestate.StringUpdate())
	}
}

func compareState(b DnsHost, d HostState) {
	if d.IsCname() != b.IsCname() {
		log.Println("CNAME mismatch can't be a CNAME and have Alias records.")
		if b.IsCname() {
			log.Printf("%v currently points to CNAME %v, can't add an IP %v", b.HostName, b.finalName, d.Target)
		} else {
			log.Printf("%v currently points to IP %v, can't add an CNAME %v", b.HostName, b.TargetIP(), d.Target)
		}
		return
	}

	if d.IsCname() {
		if d.Target != b.finalName {
			log.Printf("Need to update CNAME %v -> %v", b.finalName, d.Target)
			return
		}
		log.Printf("Nothing to do, CNAMEs already match")
		return
	}
	if !b.HasIP(d.Target) {
		log.Printf("Need to add IP %v to set %v", d.Target, b.TargetIP())
		return
	}
	log.Printf("Nothing to do, records match")
	return
}

func ProcessFlags() HostState {

	var action string
	var record string
	options := []string{"create", "update", "delete", "lookup"}
	for _, opt := range options {
		f := flag.Lookup(opt)
		if f.Value.String() != "" {
			if action != "" {
				log.Fatal("You can only specify one of ", strings.Join(options, ", "))
			}
			action = f.Name
			record = f.Value.String()

		}

	}
	t := flag.Lookup("target")
	if t.Value.String() == "" && action != "lookup" {
		log.Fatal("You must also specify one target")
	}

	var state = HostState{record, t.Value.String(), action}
	return state
}
