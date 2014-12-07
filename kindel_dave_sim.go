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
func simGate(gatetype int, invals []int) int {
	val := 2 //default of x
	switch gatetype {
	case T_or:
		val = 0
		for _, predVal := range invals {
			if predVal == 1 {
				val = 1
				break
			} else if predVal == 2 {
				val = 2
			}
		}
	case T_and:
		val = 1
		for _, predVal := range invals {
			if predVal == 0 {
				val = 0
				break
			} else if predVal == 2 {
				val = 2
			}
		}
	case T_nand:
		val = 0
		for _, predVal := range invals {
			if predVal == 0 {
				val = 1
				break
			} else if predVal == 2 {
				val = 2
			}
		}
	case T_nor:
		val = 1
		for _, predVal := range invals {
			if predVal == 1 {
				val = 0
				break
			} else if predVal == 2 {
				val = 2
			}
		}
	case T_xor:
		val = 0
		for _, predVal := range invals {
			if predVal == 1 {
				val = (val + 1) % 2 //either 0 or 1
			} else if predVal == 2 {
				val = 2
				break
			}
		}
	case T_xnor:
		val = 1
		for _, predVal := range invals {
			if predVal == 1 {
				val = (val + 1) % 2 //either 0 or 1
			} else if predVal == 2 {
				val = 2
				break
			}
		}
	case T_not:
		predVal := invals[0]
		if predVal == 2 {
			val = 2
		} else {
			val = (val + 1) % 2
		}
	case T_buf, T_output:
		val = invals[0]
	}
	return val
}
