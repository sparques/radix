package radix

func XorBEBit(buf []byte, pos int, val bool) {
	if val {
		buf[pos/8] ^= 1 << (7 - pos%8)
	}
}

func XorLEBit(buf []byte, pos int, val bool) {
	if val {
		buf[pos/8] ^= 1 << (pos % 8)
	}
}

func SetBEBit(buf []byte, pos int, val bool) {
	mask := byte(1 << (7 - pos%8))
	if val {
		buf[pos/8] |= mask
	} else {
		buf[pos/8] &^= mask
	}
}

func SetLEBit(buf []byte, pos int, val bool) {
	mask := byte(1 << (pos % 8))
	if val {
		buf[pos/8] |= mask
	} else {
		buf[pos/8] &^= mask
	}
}

func GetBEBit(buf []byte, pos int) bool {
	return (buf[pos/8]>>(7-pos%8))&1 != 0
}

func GetLEBit(buf []byte, pos int) bool {
	return (buf[pos/8]>>(pos%8))&1 != 0
}

func NRZ(bit bool) int8 {
	if bit {
		return -1
	}
	return 1
}
