package radix

// XorBEBit toggles one bit in a byte slice when val is true.
// "BE" means bit 0 is the high bit of buf[0].
func XorBEBit(buf []byte, pos int, val bool) {
	if val {
		buf[pos/8] ^= 1 << (7 - pos%8)
	}
}

// XorLEBit toggles one bit in a byte slice when val is true.
// "LE" means bit 0 is the low bit of buf[0].
func XorLEBit(buf []byte, pos int, val bool) {
	if val {
		buf[pos/8] ^= 1 << (pos % 8)
	}
}

// SetBEBit sets or clears one big-endian-numbered bit in a byte slice.
func SetBEBit(buf []byte, pos int, val bool) {
	mask := byte(1 << (7 - pos%8))
	if val {
		buf[pos/8] |= mask
	} else {
		buf[pos/8] &^= mask
	}
}

// SetLEBit sets or clears one little-endian-numbered bit in a byte slice.
func SetLEBit(buf []byte, pos int, val bool) {
	mask := byte(1 << (pos % 8))
	if val {
		buf[pos/8] |= mask
	} else {
		buf[pos/8] &^= mask
	}
}

// GetBEBit reads one big-endian-numbered bit from a byte slice.
func GetBEBit(buf []byte, pos int) bool {
	return (buf[pos/8]>>(7-pos%8))&1 != 0
}

// GetLEBit reads one little-endian-numbered bit from a byte slice.
func GetLEBit(buf []byte, pos int) bool {
	return (buf[pos/8]>>(pos%8))&1 != 0
}

// NRZ converts a Boolean bit to the modem's signed representation.
// In this package false is +1 and true is -1.
func NRZ(bit bool) int8 {
	if bit {
		return -1
	}
	return 1
}
