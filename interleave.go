package radix

import "fmt"

type interleaverParams struct {
	first  uint
	second uint
	third  uint
}

var interleavers = map[int]interleaverParams{
	8:  {1, 1, 2},
	11: {1, 3, 4},
	12: {1, 1, 4},
	13: {1, 1, 9},
	14: {1, 5, 10},
	15: {1, 1, 3},
	16: {1, 1, 14},
}

func InterleaveEncode[T any](src []T, order int) ([]T, error) {
	length := 1 << order
	if len(src) != length {
		return nil, fmt.Errorf("got %d symbols, want %d", len(src), length)
	}
	params, ok := interleavers[order]
	if !ok {
		return nil, fmt.Errorf("unsupported interleaver order %d", order)
	}
	dest := make([]T, length)
	dest[0] = src[0]
	seq := NewXorShiftMask(uint(order), params.first, params.second, params.third, 1)
	for i := 1; i < length; i++ {
		dest[i] = src[seq.Next()]
	}
	return dest, nil
}

func InterleaveDecode[T any](src []T, order int) ([]T, error) {
	length := 1 << order
	if len(src) != length {
		return nil, fmt.Errorf("got %d symbols, want %d", len(src), length)
	}
	params, ok := interleavers[order]
	if !ok {
		return nil, fmt.Errorf("unsupported interleaver order %d", order)
	}
	dest := make([]T, length)
	dest[0] = src[0]
	seq := NewXorShiftMask(uint(order), params.first, params.second, params.third, 1)
	for i := 1; i < length; i++ {
		dest[seq.Next()] = src[i]
	}
	return dest, nil
}
