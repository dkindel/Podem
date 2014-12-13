package main

import (
	"fmt"
)

type Objective struct {
	gatenum, val int
}
type PI struct {
	inputnum, val int
	alternateUsed bool
}

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

//loops through all the faults and runs podem
func runPodemAllFaults() {
	for _, fault := range ckt.faults {
		fmt.Println("----------------------------------------------------------")
		fmt.Println("Running PODEM on fault of faulty gate", fault.gatenum, "with gate type", fault.gatetype)
		ckt.gatetype2[fault.gatenum] = fault.gatetype
		//fault now injected.  Now we need to run
		if runPodem(fault) {
			fmt.Println("The vector that sensitizes faulty gate", fault.gatenum, "with gate type", fault.gatetype, "is:")
			for i := 0; i < ckt.numin; i++ {
				fmt.Println(ckt.inputs[i], " = ", ckt.value1[ckt.inputs[i]])
			}
		} else {
			fmt.Println("All possible inputs have failed.  No test is possible for faulty gate", fault.gatenum, "with gate type", fault.gatetype)
		}
	}
}

//runs podem for a single fault
func runPodem(f Fault) bool {
	//find what inputs will sensitize the fault
	inputlist := sensitizedFaultList(f)

	debugMsg("Possible gate inputs to sensitize the fault:", inputlist)

	//This is essentially running through the implication stack
	//The test can still be tested using a different input.
podemStart: //Label here so we can continue if failed
	for _, input := range inputlist {
		//select the first group of inputs that sensitize the fault
		setAllToX()

		var stack Stack
		debugMsg("Running xpath with", f.gatenum)
		for {
			for xpathCheck(f.gatenum) {
				debugMsg("Running getObjective with inputs", input, "for", f.gatenum)
				objective := getObjective(f.gatenum, input)
				debugMsg("Running a backtrace from", objective.gatenum, "with", objective.val)
				if objective.gatenum == -1 {
					panic("Uh oh.  This shouldn't have happened because the xpath should check if there's an x gate on the D frontier that we can use!")
				}
				pi := backtrace(objective)
				debugMsg("Backtrace from", objective, "provided", pi)
				stack.Push(pi)
				ckt.value1[pi.inputnum] = pi.val
				if implyAndTest() {
					return true
				}
			}
			for {
				//if we've run out of backtracking options
				if stack.Len() == 0 {
					debugMsg("Input", input, "has failed.  Moving to next input if there is any untested.")
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
					debugMsg("Backtracking. Now using pi", lastPI.inputnum, "with value", lastPI.val)
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
	return false
}

func implyAndTest() bool {
	imply()
	for i := 1; i <= ckt.numgates; i++ {
		debugMsg("gate ", i, " has logic val ", ckt.value1[i])
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

//gets the current objective.  If the faulty gate hasn't been set
//yet, it returns an X value on the faulty gate inputs
//otherwise, it searches for the D frontier and finds an X value
//on those inputs.
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
	debugMsg("xGateFromDFrontier returned ", dgate, val)
	if dgate == -1 {
		objective.gatenum = -1
		objective.val = -1
		return objective
	}
	objective.gatenum = dgate
	objective.val = val
	return objective
}

//Finds a gate that is x on the D Frontier
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

//runs the xpath check on the outputs and on the d frontier.
//we have to make sure that the test CAN be run
func xpathCheck(faultyGate int) bool {
	for _, po := range ckt.outputs {
		if xpathRecur(faultyGate, po) {
			return true
		}
	}
	//if the faulty gate isn't X, check the D frontier for an X gate
	if ckt.value1[faultyGate] != 2 {
		response, _ := xGateFromDFrontier(faultyGate, ckt.value1[faultyGate])
		if response == -1 {
			return false
		}
	}
	return true
}

//Recursive function for checking the xpath
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

//The backtrace function!
//This runs from the objective down to a PI, finding the values for
//each gate along the way.  If it encounters a not/nand/nor, it
//flips the value to set at the PI (starting with the desired
//value of the objective)
func backtrace(objective Objective) PI {
	var pi PI
	currGate := objective.gatenum
	currVal := objective.val
	for ckt.gatetype1[currGate] != T_input {
		if allInputsNeedSet(currVal, ckt.gatetype2[currGate]) {
			//follow path of hardest controllability
			numInputs := ckt.fanin[currGate]
			if currVal == 1 {
				//check for the cc1 controllability
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
			} else { //check cc0 controllability
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
			if currVal == 1 { //check cc1 controllability
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
			} else { //check cc0 controllability
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
		debugMsg("backtracing ", currGate, " with currVal ", currVal)
	}
	pi.inputnum = currGate
	pi.val = currVal
	pi.alternateUsed = false
	return pi
}

//Checks if a gate needs all of its inputs to be set to determine
//if we should follow easiest or hardest controllability
//this is determined by the desired value and the gate type
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
