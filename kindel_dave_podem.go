package main

import (
	"fmt"
	//	"reflect"
)

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
	return inputlist
}

func makeStack(slice [][]int) *Stack {
	stack := new(Stack)
	for _, subslice := range slice {
		stack.Push(subslice)
	}
	return stack
}

func removeDuplicateAttempts(slice *[][]int) {
	/*newslice := make([][]int, 0, len(slice))
	for i, _ := range newslice {
		newslice[i] = make([]int, len(slice[i]), len(slice[i]))
	}*/
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
	for inputstack.Len() > 0 {
		input := inputstack.Pop().([]int)
		setAllToX()
		//set the inputs to the gate and simulate!
		for i, val := range input {
			ckt.inlist[f.gatenum][i] = val
		}

		//we know that the value has to be either 0 or 1 here
		//anything else won't have been put into the list
		if simGate(ckt.gatetype1[f.gatenum], input) == 1 {
			ckt.value1[f.gatenum] = 3
		} else {
			ckt.value1[f.gatenum] = 4
		}

		if !justify(f.gatenum) {
			continue //failed.  Need to continue
		}
		if !prop(f.gatenum) {
			continue
		}

	}
}

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

func justify(gatenum int) bool {

	return true
}

func prop(gatenum int) bool {
	return true
}
