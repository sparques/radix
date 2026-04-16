package radix

import "testing"

func TestInterleaveRoundTrip(t *testing.T) {
	for _, order := range []int{8, 11, 12, 13, 14, 15, 16} {
		src := make([]int, 1<<order)
		for i := range src {
			src[i] = i
		}
		encoded, err := InterleaveEncode(src, order)
		if err != nil {
			t.Fatalf("InterleaveEncode order %d: %v", order, err)
		}
		decoded, err := InterleaveDecode(encoded, order)
		if err != nil {
			t.Fatalf("InterleaveDecode order %d: %v", order, err)
		}
		for i := range src {
			if decoded[i] != src[i] {
				t.Fatalf("order %d decoded[%d] = %d, want %d", order, i, decoded[i], src[i])
			}
		}
	}
}
