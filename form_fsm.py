from numba import jit, uint8

@jit(uint8(uint8, uint8))
def scan(state: int, next_char: int) -> int:
	if state == 0 and next_char == 67: state = 1
	elif state == 1 and next_char == 111: state = 2
	elif state == 2 and next_char == 110: state = 3
	elif state == 3 and next_char == 116: state = 4
	elif state == 4 and next_char == 101: state = 5
	elif state == 5 and next_char == 110: state = 6
	elif state == 6 and next_char == 116: state = 7
	elif state == 7 and next_char == 45: state = 8
	elif state == 8 and next_char == 84: state = 9
	elif state == 9 and next_char == 121: state = 10
	elif state == 10 and next_char == 112: state = 11
	return state
