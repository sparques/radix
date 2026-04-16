package radix

import "fmt"

func PolarEncode(message []int8, frozen []uint32, order int) ([]int8, error) {
	length := 1 << order
	if len(frozen)*32 < length {
		return nil, fmt.Errorf("frozen table has %d bits, want %d", len(frozen)*32, length)
	}

	codeword := make([]int8, length)
	msgIdx := 0
	for i := 0; i < length; i += 2 {
		msg0 := int8(1)
		if !frozenBit(frozen, i) {
			if msgIdx >= len(message) {
				return nil, fmt.Errorf("message ended at bit %d", msgIdx)
			}
			msg0 = message[msgIdx]
			msgIdx++
		}
		msg1 := int8(1)
		if !frozenBit(frozen, i+1) {
			if msgIdx >= len(message) {
				return nil, fmt.Errorf("message ended at bit %d", msgIdx)
			}
			msg1 = message[msgIdx]
			msgIdx++
		}
		codeword[i] = msg0 * msg1
		codeword[i+1] = msg1
	}
	if msgIdx != len(message) {
		return nil, fmt.Errorf("message has %d unused bits", len(message)-msgIdx)
	}

	for h := 2; h < length; h *= 2 {
		for i := 0; i < length; i += 2 * h {
			for j := i; j < i+h; j++ {
				codeword[j] *= codeword[j+h]
			}
		}
	}
	return codeword, nil
}

func frozenBit(bits []uint32, idx int) bool {
	return (bits[idx/32]>>(idx%32))&1 != 0
}
