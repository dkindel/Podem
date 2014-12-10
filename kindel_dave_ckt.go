package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	JUNK = iota
	T_input
	T_output
	T_xor
	T_xnor
	T_dff
	T_and
	T_nand
	T_or
	T_nor
	T_not
	T_buf
)

const MAXLEVELS = 50

type Circuit struct {
	numin       int
	numout      int
	numgates    int
	numlevels   int
	numfaults   int
	value1      []int //0 is 0, 1 is 1, 2 is X, 3 is D, 4 is !D
	value2      []int //value 2 is the faulty value
	inlist      [][]int
	outlist     [][]int
	gateByLevel [][]int
	gatetype1   []int
	gatetype2   []int
	levelNum    []int
	fanin       []int
	fanout      []int
	inputs      []int
	outputs     []int
	faults      []Fault
	stacks      []Stack //each stack consists of possible inputs left
}

type Fault struct {
	gatenum  int
	gatetype int
}

//builos the circuit and loads it into the global ckt var
func makecircuit(cktname string) {
	file, err := os.Open(cktname + ".lev")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//ckt is defined in the main go file
	// ckt := new(Circuit)
	var count, junk, netnum, numpri, numout int
	fmt.Fscanf(file, "%d", &count)
	fmt.Fscanf(file, "%d", &junk)
	ckt.numgates = count - 1
	ckt.numlevels = 0

	ckt.gatetype1 = make([]int, count)
	ckt.gatetype2 = make([]int, count)
	ckt.value1 = make([]int, count)
	ckt.value2 = make([]int, count)
	ckt.levelNum = make([]int, count)
	ckt.fanin = make([]int, count)
	ckt.fanout = make([]int, count)
	ckt.inputs = make([]int, count)
	ckt.outputs = make([]int, count)
	ckt.stacks = make([]Stack, count)

	ckt.inlist = make([][]int, count)
	ckt.outlist = make([][]int, count)

	//initialize the gateByLevel (want to be able to append)
	ckt.gateByLevel = make([][]int, MAXLEVELS)
	for i := 0; i < MAXLEVELS; i++ {
		ckt.gateByLevel[i] = make([]int, 0, count)
	}

	//use scanner to scan line by line
	scanner := bufio.NewScanner(file)

	for i := 1; i < count; i++ {
		//get the line text
		scanner.Scan()
		txt := scanner.Text()

		//split the line by a space
		str_split := strings.Split(txt, " ")
		split := make([]int, len(str_split))
		for j := 0; j < len(str_split); j++ {
			split_int, err := strconv.Atoi(str_split[j])
			if err != nil {
				continue
			}
			split[j] = split_int
		}

		split_pos := 0

		netnum = split[0]
		ckt.gatetype1[netnum] = split[1]
		ckt.gatetype2[netnum] = ckt.gatetype1[netnum]
		ckt.levelNum[netnum] = split[2]
		ckt.fanin[netnum] = split[3]
		split_pos += 4

		//always start with every gate as X
		ckt.value1[netnum] = 2
		ckt.value2[netnum] = 2

		if ckt.levelNum[netnum] > ckt.numlevels-1 {
			ckt.numlevels = ckt.levelNum[netnum] + 1 //the number of levels is that num+1
		}

		//build levels slice
		ckt.gateByLevel[ckt.levelNum[netnum]] = append(ckt.gateByLevel[ckt.levelNum[netnum]], netnum)

		//build input slice
		if ckt.gatetype1[netnum] == T_input {
			ckt.inputs[numpri] = netnum
			numpri++
		}

		//build fanin list slice
		ckt.inlist[netnum] = make([]int, ckt.fanin[netnum])
		for j := 0; j < ckt.fanin[netnum]; j++ {
			ckt.inlist[netnum][j] = split[split_pos]
			split_pos++
		}

		split_pos += ckt.fanin[netnum]

		//build fanout slice
		ckt.fanout[netnum] = split[split_pos]
		split_pos++

		if ckt.gatetype1[netnum] == T_output {
			ckt.outputs[numout] = netnum
			numout++
		}

		//build fanout list slice
		ckt.outlist[netnum] = make([]int, ckt.fanout[netnum])
		for j := 0; j < ckt.fanout[netnum]; j++ {
			ckt.outlist[netnum][j] = split[split_pos]
			split_pos++
		}

	}
	// return ckt
}

//applys an input vector to the circuit, performs a gate level simulation of it
//and then prints out the values
func logicSimFromFile(cktname string) {
	file, err := os.Open(cktname + ".vec")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	vecWidth_txt := scanner.Text()
	vecWidth, _ := strconv.Atoi(vecWidth_txt)

	for scanner.Scan() {
		txt := scanner.Text()
		vect := translateVector(txt, vecWidth)
		applyVector(vect)
		goodsim()
		printValByLevel()
	}
}

//translates a vector string to a int slice
func translateVector(vec string, vecWidth int) []int {
	fmt.Println(vec)
	vec_slice := make([]int, vecWidth)
	for i, c := range vec {
		char := string(c)
		var val = 0
		//case insensitive comparison
		if strings.EqualFold(char, "x") {
			val = 2
		} else {
			val, _ = strconv.Atoi(char)
		}
		vec_slice[i] = val
	}
	return vec_slice
}

//applies the vector to the PI's
func applyVector(vect []int) {
	for i, inval := range vect {
		ckt.value1[ckt.inputs[i]] = inval
		ckt.value2[ckt.inputs[i]] = inval
	}
}

//loads the faults specified in the
func loadFaults(cktname string) {
	file, err := os.Open(cktname + ".flt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	ckt.numfaults, _ = strconv.Atoi(scanner.Text())
	ckt.faults = make([]Fault, 0, ckt.numfaults)

	for scanner.Scan() {
		txt := scanner.Text()
		f := translateFaultLine(txt)
		ckt.faults = append(ckt.faults, f)
	}
}

//translates a line of a fault text file into the Fault struct
func translateFaultLine(fault string) Fault {
	var f Fault
	//split the line by a space
	str_split := strings.Split(fault, " ")
	f.gatenum, _ = strconv.Atoi(str_split[0])
	f.gatetype, _ = strconv.Atoi(str_split[1])
	return f
}

//Helper function, in case you want to view the level layout of the gates
func printLevels() {
	for i := 0; i < ckt.numlevels; i++ {
		fmt.Printf("Level %d\n", i)
		for j := 0; j < len(ckt.gateByLevel[i]); j++ {
			fmt.Printf("%d ", ckt.gateByLevel[i][j])
		}
		fmt.Println("")
	}
}

//Helper function, in case you want to view the level layout of the gates
func printValByLevel() {
	for i := 0; i < ckt.numlevels; i++ {
		fmt.Printf("Level %d\n", i)
		for j := 0; j < len(ckt.gateByLevel[i]); j++ {
			fmt.Printf("%d ", ckt.value1[ckt.gateByLevel[i][j]])
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func intpow(a, b int) int {
	p := 1
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}
