package main

import (
	"os"
)

var ckt Circuit

func main() {
	if len(os.Args) != 2 {
		panic("No argument provided!")
	}
	name := os.Args[1]

	//builds the circuit ckt
	makecircuit(name)
	/*for i := 1; i <= ckt.numgates; i++ {
		fmt.Printf("%d %d %d %d ", i, ckt.gatetype1[i], ckt.levelNum[i], ckt.fanin[i])
		for j := 0; j < ckt.fanin[i]; j++ {
			fmt.Printf("%d ", ckt.inlist[i][j])
		}
		fmt.Printf("%d ", ckt.fanout[i])
		for j := 0; j < ckt.fanout[i]; j++ {
			fmt.Printf("%d ", ckt.outlist[i][j])
		}
		fmt.Println("")
	}
	logicSimFromFile(name)*/
	loadFaults(name) //load the faults from the file
	runPodemAllFaults()
	//makeInputList(8, 1)
}
