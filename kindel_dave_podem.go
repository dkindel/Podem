package main

import (
	"fmt"
	"sort"
)

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
			currFaninGate := ckt.inlist[gateNum][j]
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
func runPodem(f Fault) {
	//find what inputs will sensitize the fault
	inputlist := sensitizedFaultList(f)
	inputstack := makeStack(inputlist)

podemStart: //marks the start of podem.  Used to break out of a loop and
	//start with the next value if a justification is impossible
	for inputstack.Len() > 0 {
		input := inputstack.Pop().([]int)
		//fmt.Println("justifying for ", input)
		setAllToX()
		clearAllStacks()
		//set the inputs to the gate and simulate!
		for i, val := range input {
			ckt.value1[ckt.inlist[f.gatenum][i]] = val
		}

		//we know that the value has to be either 0 or 1 here
		//anything else won't have been put into the list
		if simGate(ckt.gatetype1[f.gatenum], input) == 1 {
			ckt.value1[f.gatenum] = 3
		} else {
			ckt.value1[f.gatenum] = 4
		}
		//fmt.Println(ckt.inlist[f.gatenum])
		for _, inputNum := range ckt.inlist[f.gatenum] {
			//fmt.Println("running justify for ", inputNum, " under ", f.gatenum)
			if !justify(inputNum) {
				fmt.Println("podem failure")
				continue podemStart //failed.  Need to continue
			}
		}
		if !prop(f.gatenum) {
			continue
		}
		fmt.Println("success")
	}
	for i := 1; i <= ckt.numgates; i++ {
		fmt.Println("gate ", i, " has logic val ", ckt.value1[i])
	}
}

//sets all the values in the entire circuit to x
func setAllToX() {
	for i := 1; i <= ckt.numgates; i++ {
		ckt.value1[i] = 2
		ckt.value2[i] = 2
	}
}

//clears all of the stacks in the circuit
func clearAllStacks() {
	for i := 1; i <= ckt.numgates; i++ {
		for ckt.stacks[i].Len() > 0 {
			ckt.stacks[i].Pop()
		}
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

//justify function for podem! This function completes
//the first "half" of the algorithm
func justify(gatenum int) bool {
	//fmt.Println("running justify on ", gatenum)
	//check if it's an input or if the value of the gate you're
	//trying to justfiy is an X.  If it is, it's already justified
	if ckt.gatetype1[gatenum] == T_input || ckt.value1[gatenum] == 2 {
		return true
	}
	//get possible values
	filledInputVals := makeInputList(gatenum, ckt.value1[gatenum])
	for _, item := range filledInputVals {
		ckt.stacks[gatenum].Push(item)
	}
	justified := false

	savedInputState := make([]int, ckt.fanin[gatenum])
	for i, val := range ckt.inlist[gatenum] {
		savedInputState[i] = ckt.value1[val]
	}

	for ckt.stacks[gatenum].Len() > 0 {
		list := ckt.stacks[gatenum].Pop().([]int)
		for i, val := range list {
			if val == 2 { //the value x will already be justified
				justified = true
				continue
			}
			inputnum := ckt.inlist[gatenum][i]
			//if the value already in the ckt isn't x,
			//we need to check if it works for us
			if ckt.value1[inputnum] != 2 {
				if ckt.value1[inputnum] == val { //this value works
					continue
				} else {
					justified = false
					break
				}
			}
			ckt.value1[inputnum] = val
			justified = justify(inputnum)
			if !justified {
				//one of the inputs couldn't be justified
				//invalidate every input and start over
				for i, val := range ckt.inlist[gatenum] {
					ckt.value1[val] = savedInputState[i]
				}
				break
			}
		}
		if justified {
			return true
		}
	}
	return justified
}

//Propogate function for podem.  This function
//completes the second "half" of podem
func prop(gatenum int) bool {
	for _, gate := range ckt.outlist[gatenum] {
		ckt.value1[gate] = 4
		justify(gate)
	}

	return true
}
