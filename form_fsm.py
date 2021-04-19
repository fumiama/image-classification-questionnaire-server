from numba import jit, int32

@jit
def scan(state, next_char):
	if state == 0 and next_char == b'C': state += 1
	elif state == 1 and next_char == b'o': state += 1
	elif state == 2 and next_char == b'n': state += 1
	elif state == 3 and next_char == b't': state += 1
	elif state == 4 and next_char == b'e': state += 1
	elif state == 5 and next_char == b'n': state += 1
	elif state == 6 and next_char == b't': state += 1
	elif state == 7 and next_char == b'-': state += 1
	elif state == 8 and next_char == b'T': state += 1
	elif state == 9 and next_char == b'y': state += 1
	elif state == 10 and next_char == b'p': state += 1
	return state
