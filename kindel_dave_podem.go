package main

import (
	"fmt"
	"sort"
)

type Objective struct {
	gatenum, val int
}
type PI struct {
	inputnum, val int
	alternateUsed bool
}

type ByInputX [][]int

//sort functions to sort by x
//This helps keeps x's in the circuit as long as possible
func (a ByInputX) Len() int {
	return len(a)
}

func (a ByInputX) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByInputX) Less(i, j int) bool {
	numXinI, numXinJ := 0, 0
	for _, val := range a[i] {
		if val == 2 {
			numXinI++
		}
	}
	for _, val := range a[j] {
		if val == 2 {
			numXinJ++
		}
	}
	return numXinI < numXinJ
}

//end sort functions

//gets a list of the inputs that will sensitize the fault
func sensitizedFaultList(fault Fault) [][]int {
	pow := intpow(2, ckt.fanin[fault.gatenum])
	inputlist := make([][]int, 0, pow)
	//make a slice in the loop below and appened whenever needed

	for i := 0; i < intpow(2, ckt.fanin[fault.gatenum]); i++ {
		attempt := make([]int, ckt.fanin[fault.gatenum])
		for j := 0; j < ckt.fanin[fault.gatenum]; j++ {
			pow := intpow(2, j)
			pow = pow & i
			var bit int
			if pow != 0 {
				bit = 1
			} else {
				bit = 0
			}
			attempt[ckt.fanin[fault.gatenum]-1-j] = bit
		}
		realval := simGate(ckt.gatetype1[fault.gatenum], attempt)
		faultval := simGate(ckt.gatetype2[fault.gatenum], attempt)
		if realval != faultval {
			inputlist = append(inputlist, attempt)
		}
	}
	return inputlist
}

//similar to finding the list that sensitizes the fault but
//this function finds all inputs (by only changing x vals)
//that make the outval happen
func makeInputList(gateNum, outVal int) [][]int {
	//just to start, check to see if the output is a DC
	//if it is, we know that the input list is all X's as well
	if outVal == 2 {
		inputlist := make([][]int, 1)
		inputlist[0] = make([]int, ckt.fanin[gateNum])
		for i := 0; i < ckt.fanin[gateNum]; i++ {
			inputlist[0][i] = 2
		}
		return inputlist
	}

	numVals := 3 //this lets us know how many variables this function genereates for
	//2 is only 0 and 1, 3 is 0, 1, and X, 5 is all values including D's
	pow := intpow(numVals, ckt.fanin[gateNum])
	inputlist := make([][]int, 0, pow)
	allcombinations := make([][]int, 0, pow)

	//ckt.value1[7] = 3

	for i := 0; i < pow; i++ {
		attempt := make([]int, ckt.fanin[gateNum])
		index := 0
		for j := ckt.fanin[gateNum] - 1; j >= 0; j-- {
			currFaninGate := ckt.cc0inlist[gateNum][j]
			if ckt.value1[currFaninGate] == 2 { //only change if it's x
				div := intpow(numVals, index)
				attempt[j] = (i / div) % numVals
			} else {
				attempt[j] = ckt.value1[currFaninGate]
			}
			index++
		}
		allcombinations = append(allcombinations, attempt)
	}

	removeDuplicateAttempts(&allcombinations)
	//after removing the duplicates, we now have all the possible combinations
	//of inputs.  We now need to choose only the inputs which will provide us
	//with an out put of outVal.
	//Remember that in building the combination list, only values with x have
	//been changed here since anything else has been changed in other places
	//fmt.Println(allcombinations)
	for _, slice := range allcombinations {
		val := simGate(ckt.gatetype1[gateNum], slice)
		if val == outVal {
			inputlist = append(inputlist, slice)
		}
	}

	/*fmt.Println(inputlist)
	stack := makeStack(inputlist)
	for stack.Len() > 0 {
		slice := stack.Pop()
		fmt.Println(slice)
	}*/
	sort.Sort(ByInputX(inputlist))
	return inputlist
}

func makeStack(slice [][]int) *Stack {
	stack := new(Stack)
	for _, subslice := range slice {
		stack.Push(subslice)
	}
	return stack
}

//removes all duplicates in the slice
//this is essential because many wasted attempts can
//be performed on inputs that have already been tried
func removeDuplicateAttempts(slice *[][]int) {
	length := len(*slice) - 1
	for i := 0; i < length; i++ {
		for j := i + 1; j <= length; j++ {
			realslice := *slice
			if slicesAreEqual(realslice[i], realslice[j]) {
				realslice[j] = realslice[length]
				realslice = realslice[0:length]
				*slice = realslice
				length--
				j--
			}
		}
	}
}

//Tests of 2 slices are equal
//just a helper function that's used
func slicesAreEqual(slice1, slice2 []int) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, item := range slice1 {
		if item != slice2[i] {
			return false
		}
	}
	return true
}

//loops through all the faults and runs podem
func runPodemAllFaults() {
	for _, fault := range ckt.faults {
		ckt.gatetype2[fault.gatenum] = fault.gatetype
		//fault now injected.  Now we need to run
		runPodem(fault)
	}
}

//runs podem for a single fault
func runPodem(f Fault) bool {
	//find what inputs will sensitize the fault
	inputlist := sensitizedFaultList(f)

	varDseen := false
	varDBseen := false

	fmt.Println(inputlist)

	//This is essentially running through the implication stack
	//The test can still be tested using a different input.
podemStart: //Label here so we can continue if failed
	for _, input := range inputlist {
		//select the first group of inputs that sensitize the fault
		setAllToX()
		if simGate(ckt.gatetype1[f.gatenum], input) == 1 {
			if varDseen { //no need to run podem for the same variable
				continue
			}
			varDseen = true
		} else {
			if varDBseen {
				continue
			}
			varDBseen = true
		}

		var stack Stack
		fmt.Println("running xpath with ", f.gatenum)
		for {
			for xpathCheck(f.gatenum) {
				fmt.Println("running getObjective with inputs ", input, " for ", f.gatenum)
				objective := getObjective(f.gatenum, input)
				fmt.Println("getObjective returned with ", objective)
				if objective.gatenum == -1 {
					panic("uh oh")
				}
				pi := backtrace(objective)
				fmt.Println("backtrace from ", objective, " provided ", pi)
				stack.Push(pi)
				ckt.value1[pi.inputnum] = pi.val
				if implyAndTest() {
					return true
				}
			}
			for {
				//if we've run out of backtracking options
				if stack.Len() == 0 {
					fmt.Println("input ", input, " has failed.  Moving to next input.")
					continue podemStart
				}
				//otherwise, we can backtrack!
				lastPI := stack.Pop().(PI)
				if !lastPI.alternateUsed {
					lastPI.alternateUsed = true
					if lastPI.val == 0 {
						lastPI.val = 1
					} else {
						lastPI.val = 0
					}
					stack.Push(lastPI)
					ckt.value1[lastPI.inputnum] = lastPI.val
					fmt.Println("Backtracking. Now using pi ", lastPI.inputnum, " with value ", lastPI.val)
					if implyAndTest() {
						return true
					}
					break //perform xtest again and continue
				}
			}
		}
	}
	/*
		for i := 1; i <= ckt.numgates; i++ {
			fmt.Println("gate ", i, " has logic val ", ckt.value1[i])
		}*/
	fmt.Println("All possible inputs have failed.  No test is possible.")
	return false
}

func implyAndTest() bool {
	imply()
	for i := 1; i <= ckt.numgates; i++ {
		fmt.Println("gate ", i, " has logic val ", ckt.value1[i])
	}
	for _, po := range ckt.outputs {
		if ckt.value1[po] == 3 || ckt.value1[po] == 4 {
			fmt.Println("SUCCESS! D has been propogated")
			return true
		}
	}
	return false
}

//sets all the values in the entire circuit to x
func setAllToX() {
	for i := 1; i <= ckt.numgates; i++ {
		ckt.value1[i] = 2
		ckt.value2[i] = 2
	}
}

//just a helper function to print off the possible input values
//that'll sensitize the list
func runSensList() {
	for _, f := range ckt.faults {
		inputlist := sensitizedFaultList(f)
		fmt.Println(inputlist)
	}
}

func getObjective(gatenum int, faultGateInputs []int) Objective {
	var objective Objective
	for i, inputval := range faultGateInputs {
		//get the input gate i of the faulty gate, gatenum
		inputgate := ckt.cc0inlist[gatenum][i]
		//if the input i of the fault gate, gatenum, is x
		if ckt.value1[inputgate] == 2 {
			//set the objective and val to the
			objective.gatenum = inputgate
			objective.val = inputval
			return objective
		}
	}
	dgate, val := xGateFromDFrontier(gatenum, ckt.value1[gatenum])
	fmt.Println("xGateFromDFrontier returned ", dgate, val)
	if dgate == -1 {
		objective.gatenum = -1
		objective.val = -1
		return objective
	}
	/*if dgate == gatenum {
		objective.gatenum = gatenum
		if val == 1 {
			objective.val = 0
		}
		if val == 0 {
			objective.val = 1
		}
		return objective
	}
	c := 0
	controlgate, controlval := controllingInput(dgate)
	if controlgate != -1 {
		if controlval != 2 {
			c = controlval
		}
		dgate = controlgate
	}
	objective.gatenum = dgate
	if c == 0 {
		objective.val = 1
	} else {
		objective.val = 0
	}*/
	objective.gatenum = dgate
	objective.val = val
	return objective
}

func controllingInput(dgate int) (int, int) {
	val := simGate(ckt.gatetype2[dgate], ckt.cc0inlist[dgate])
	switch ckt.gatetype2[dgate] {
	case T_or:
		fallthrough
	case T_nand:
		if val == 0 {
			return ckt.cc0inlist[dgate][0], 0
		} else if val == 1 {
			numOnes := 0
			var gate int
			for _, input := range ckt.cc0inlist[dgate] {
				if ckt.value1[input] == 1 {
					numOnes++
					gate = input
				}
			}
			if numOnes == 1 {
				return gate, 1
			}
		} else if val == 2 {
			numX := 0
			var gate int
			for _, input := range ckt.cc0inlist[dgate] {
				if ckt.value1[input] == 2 {
					numX++
					gate = input
				}
			}
			if numX == 1 {
				return gate, 2
			}
		}
	case T_nor:
		fallthrough
	case T_and:
		if val == 1 {
			return ckt.cc0inlist[dgate][0], 1
		} else if val == 0 {
			numZeroes := 0
			var gate int
			for _, input := range ckt.cc0inlist[dgate] {
				if ckt.value1[input] == 0 {
					numZeroes++
					gate = input
				}
			}
			if numZeroes == 1 {
				return gate, 0
			}
		} else if val == 2 {
			numX := 0
			var gate int
			for _, input := range ckt.cc0inlist[dgate] {
				if ckt.value1[input] == 2 {
					numX++
					gate = input
				}
			}
			if numX == 1 {
				return gate, 2
			}
		}
	}
	return -1, -1 //no controlling gate found
}

func xGateFromDFrontier(gatenum, val int) (int, int) {
	for _, output := range ckt.outlist[gatenum] {
		if (ckt.value1[output] == 3) || (ckt.value1[output] == 4) {
			return xGateFromDFrontier(output, ckt.value1[output])
		}
	}
	for _, output := range ckt.outlist[gatenum] {
		if ckt.value1[output] == 2 {
			gtype := ckt.gatetype2[output]
			flipVal := (gtype == T_not) || (gtype == T_nand) || (gtype == T_nor)
			if val == 3 {
				if flipVal {
					return output, 4
				}
				return output, 3
			} else {
				if flipVal {
					return output, 3
				}
				return output, 4
			}
		}
	}
	return -1, -1
}

func xpathCheck(faultyGate int) bool {
	for _, po := range ckt.outputs {
		if xpathRecur(faultyGate, po) {
			return true
		}
	}
	if ckt.value1[faultyGate] != 2 {
		response, _ := xGateFromDFrontier(faultyGate, ckt.value1[faultyGate])
		if response == -1 {
			return false
		}
	}
	return true
}

func xpathRecur(faultyGate, gate int) bool {
	for _, input := range ckt.cc0inlist[gate] {
		//we're looking at the faulty gate.
		if input == faultyGate {
			//if the faulty gate is an x, D can still be progated
			if ckt.value1[input] == 2 {
				return true
			} else { //if it's not an x, we have to backtrack
				return false
			}
		}
		//looking at a different gate
		if (ckt.value1[input] == 3) || (ckt.value1[input] == 4) {
			return true
		} else if ckt.value1[input] == 2 {
			return xpathRecur(faultyGate, input)
		}
	}
	return false
}

func backtrace(objective Objective) PI {
	var pi PI
	currGate := objective.gatenum
	currVal := objective.val
	for ckt.gatetype1[currGate] != T_input {
		if allInputsNeedSet(currVal, ckt.gatetype2[currGate]) {
			//follow path of hardest controllability
			numInputs := ckt.fanin[currGate]
			if currVal == 1 {
				for i := numInputs - 1; i >= 0; i-- {
					input := ckt.cc1inlist[currGate][i]
					if ckt.value1[input] == 2 {
						gtype := ckt.gatetype2[currGate]
						flipVal := (gtype == T_not) || (gtype == T_nand) || (gtype == T_nor)
						if flipVal {
							currVal = 0
						}
						currGate = input
						break
					}
				}
			} else {
				for i := numInputs - 1; i >= 0; i-- {
					input := ckt.cc0inlist[currGate][i]
					if ckt.value1[input] == 2 {
						gtype := ckt.gatetype2[currGate]
						flipVal := (gtype == T_not) || (gtype == T_nand) || (gtype == T_nor)
						if flipVal {
							currVal = 1
						}
						currGate = input
						break
					}
				}

			}
		} else {
			//follow easiest controllability
			numInputs := ckt.fanin[currGate]
			if currVal == 1 {
				for i := 0; i < numInputs; i++ {
					input := ckt.cc1inlist[currGate][i]
					if ckt.value1[input] == 2 {
						gtype := ckt.gatetype2[currGate]
						flipVal := (gtype == T_not) || (gtype == T_nand) || (gtype == T_nor)
						if flipVal {
							currVal = 0
						}
						currGate = input
						break
					}
				}
			} else {
				for i := 0; i < numInputs; i++ {
					input := ckt.cc0inlist[currGate][i]
					if ckt.value1[input] == 2 {
						gtype := ckt.gatetype2[currGate]
						flipVal := (gtype == T_not) || (gtype == T_nand) || (gtype == T_nor)
						if flipVal {
							currVal = 1
						}
						currGate = input
						break
					}
				}
			}
		}
		fmt.Println("backtracing ", currGate, " with currVal ", currVal)
	}
	pi.inputnum = currGate
	pi.val = currVal
	pi.alternateUsed = false
	return pi
}

func allInputsNeedSet(gateval, gatetype int) bool {
	if gateval == 0 {
		switch gatetype {
		case T_not:
			fallthrough
		case T_buf:
			fallthrough
		case T_or:
			fallthrough
		case T_xor:
			fallthrough
		case T_xnor:
			fallthrough
		case T_nand:
			return true
		default:
			return false
		}
	} else {
		switch gatetype {
		case T_not:
			fallthrough
		case T_buf:
			fallthrough
		case T_nor:
			fallthrough
		case T_and:
			fallthrough
		case T_xor:
			fallthrough
		case T_xnor:
			return true
		default:
			return false
		}
	}
}
