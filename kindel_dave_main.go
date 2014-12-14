package main

import (
	"fmt"
	"os"
	"strings"
)

var ckt Circuit
var debugOn bool

func main() {
	//find the file name and check for the debug flag
	if len(os.Args) < 2 {
		panic("No argument provided!")
	} else if len(os.Args) >= 4 {
		panic("Too many arguments provided!")
	}
	name := os.Args[1]
	if len(os.Args) == 3 {
		debugOn = strings.EqualFold(os.Args[2], "-debug")
	} else {
		debugOn = false
	}

	//builds the circuit ckt
	makecircuit(name)

	loadFaults(name) //load the faults from the file
	faultSuccesses, faultFailures := runPodemAllFaults()

	fmt.Println(faultFailures, "faults are undetectable")
	fmt.Println(faultSuccesses, "faults have been successfully simulated")
	//makeInputList(8, 1)
}

func debugMsg(a ...interface{}) {
	if debugOn {
		fmt.Println(a...)
	}
}
