package radix

import "math/bits"

func HadamardEncode7(message int) [SeedTones]int8 {
	var code [SeedTones]int8
	for i := range code {
		if bits.OnesCount(uint(message&(i|SeedTones)))%2 == 0 {
			code[i] = 1
		} else {
			code[i] = -1
		}
	}
	return code
}

func HadamardDecode7(code []int8) int {
	if len(code) != SeedTones {
		return -1
	}

	sum := make([]int, SeedTones)
	for i := 0; i < SeedTones-1; i += 2 {
		sum[i] = int(code[i]) + int(code[i+1])
		sum[i+1] = int(code[i]) - int(code[i+1])
	}
	for h := 2; h < SeedTones; h *= 2 {
		for i := 0; i < SeedTones; i += 2 * h {
			for j := i; j < i+h; j++ {
				x := sum[j] + sum[j+h]
				y := sum[j] - sum[j+h]
				sum[j] = x
				sum[j+h] = y
			}
		}
	}

	word, best, next := 0, 0, 0
	for i := 0; i < SeedTones; i++ {
		mag := absInt(sum[i])
		msg := i
		if sum[i] < 0 {
			msg += SeedTones
		}
		if mag > best {
			next = best
			best = mag
			word = msg
		} else if mag > next {
			next = mag
		}
	}
	if best == next {
		return -1
	}
	return word
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
