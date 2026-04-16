package radix

import "fmt"

// PolarEncode applies the Radix polar code to signed message symbols.
// Symbols use the package convention +1/-1 instead of 0/1 bits.
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

// PolarDecodeHard decodes a polar codeword using hard decisions only.
// This is suitable for clean captures; noisy real-world decoding will eventually
// want a soft/list decoder.
func PolarDecodeHard(codeword []int8, frozen []uint32, order int) ([]int8, error) {
	length := 1 << order
	if len(codeword) != length {
		return nil, fmt.Errorf("got %d codeword symbols, want %d", len(codeword), length)
	}
	if len(frozen)*32 < length {
		return nil, fmt.Errorf("frozen table has %d bits, want %d", len(frozen)*32, length)
	}

	u := append([]int8(nil), codeword...)
	polarTransform(u)
	message := make([]int8, 0, countInformationBits(frozen, order))
	for i, symbol := range u {
		if frozenBit(frozen, i) {
			continue
		}
		message = append(message, signInt8(symbol))
	}
	return message, nil
}

func polarTransform(symbols []int8) {
	length := len(symbols)
	for i := 0; i < length; i += 2 {
		symbols[i] *= symbols[i+1]
	}
	for h := 2; h < length; h *= 2 {
		for i := 0; i < length; i += 2 * h {
			for j := i; j < i+h; j++ {
				symbols[j] *= symbols[j+h]
			}
		}
	}
}

func frozenBit(bits []uint32, idx int) bool {
	return (bits[idx/32]>>(idx%32))&1 != 0
}

func signInt8(v int8) int8 {
	if v < 0 {
		return -1
	}
	return 1
}
