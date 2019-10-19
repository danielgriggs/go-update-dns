package main

import "regexp"

type HostState struct {
	Record string
	Target string
	Action string
}

func (h HostState) IsCname() bool {
	match, _ := regexp.MatchString(`^[\d\.]*$`, h.Target)
	if match {
		return false
	} else {
		return true
	}
}
