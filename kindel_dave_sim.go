package main

func goodsim() {
	//first level is already filled in
	//loop through the rest
	for i := 1; i < ckt.numlevels; i++ {
		for j := 0; j < len(ckt.gateByLevel[i]); j++ {
			gatenum := ckt.gateByLevel[i][j]
			faninslice := ckt.inlist[gatenum]

			faninVal := make([]int, len(faninslice))
			for k, gate := range faninslice {
				faninVal[k] = ckt.value1[gate]
			}
			ckt.value1[gatenum] = simGate(ckt.gatetype1[gatenum], faninVal)
		}
	}
}

//This will simulate a gate of type gatetype using the invals as the fanin list
//the invals list doesn't care about anything BUT the values
//and doesn't necessarily correspond to the circuit at all
//for obvious reasons, this doesn't work with inputs
//
//This takes D and !D into account as well
func simGate(gatetype int, invals []int) int {
	val := 2 //default of x
	switch gatetype {
	case T_or:
		val = 0
		seenD := false  //used to tell if a D has been seen
		seenDB := false //used to tell if a !D has been seen
		for _, predVal := range invals {
			//if there's ever a 0 or a conflicting D or !D on the inputs
			if predVal == 1 || (seenD && predVal == 4) || (seenDB && predVal == 3) {
				val = 1
				break
			} else if predVal == 2 {
				val = 2
			} else if val != 2 {
				if (val == 0 || val == 3) && predVal == 3 {
					val = 3
				} else if (val == 0 || val == 4) && predVal == 4 {
					val = 4
				}
			}
			if predVal == 3 {
				seenD = true
			} else if predVal == 4 {
				seenDB = true
			}
		}
	case T_and:
		val = 1
		seenD := false  //used to tell if a D has been seen
		seenDB := false //used to tell if a !D has been seen
		for _, predVal := range invals {
			//if there's ever a 0 or a conflicting D or !D on the inputs
			if predVal == 0 || (seenD && predVal == 4) || (seenDB && predVal == 3) {
				val = 0
				break
			} else if predVal == 2 {
				val = 2
			} else if val != 2 {
				if (val == 1 || val == 3) && predVal == 3 {
					val = 3
				} else if (val == 1 || val == 4) && predVal == 4 {
					val = 4
				}
			}
			if predVal == 3 {
				seenD = true
			} else if predVal == 4 {
				seenDB = true
			}
		}
	case T_nand:
		val = simGate(T_and, invals)
		if val == 1 {
			val = 0
		} else if val == 0 {
			val = 1
		} else if val == 3 {
			val = 4
		} else if val == 4 {
			val = 3
		}
	case T_nor:
		val = simGate(T_or, invals)
		if val == 1 {
			val = 0
		} else if val == 0 {
			val = 1
		} else if val == 3 {
			val = 4
		} else if val == 4 {
			val = 3
		}
	case T_xor:
		val = 0
		for _, predVal := range invals {
			if predVal == 2 {
				val = 2
				break
			} else if val == 0 && predVal == 1 {
				val = 1
			} else if val == 1 && predVal == 1 {
				val = 0
			} else if (val == 3 && predVal == 3) || (val == 4 && predVal == 4) {
				val = 0
			} else if val == 0 && (predVal == 3 || predVal == 4) {
				val = predVal
			} else if (val == 1 && predVal == 4) || (val == 4 && predVal == 1) {
				val = 3
			} else if (val == 1 && predVal == 3) || (val == 3 && predVal == 1) {
				val = 4
			} else if (val == 3 && predVal == 4) || (val == 4 && predVal == 3) {
				val = 1
			}
		}
	case T_xnor:
		val = simGate(T_xor, invals)
		if val == 1 {
			val = 0
		} else if val == 0 {
			val = 1
		} else if val == 3 {
			val = 4
		} else if val == 4 {
			val = 3
		}

	case T_not:
		predVal := invals[0]
		if predVal == 2 {
			val = 2
		} else if predVal == 1 {
			val = 0
		} else if predVal == 0 {
			val = 1
		} else if predVal == 3 {
			val = 4
		} else if predVal == 4 {
			val = 3
		}
	case T_buf, T_output:
		val = invals[0]
	}
	return val
}
