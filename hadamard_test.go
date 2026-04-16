package radix

import "testing"

func TestHadamard7RoundTrip(t *testing.T) {
	for msg := 0; msg < 128; msg++ {
		code := HadamardEncode7(msg)
		if got := HadamardDecode7(code[:]); got != msg {
			t.Fatalf("HadamardDecode7(HadamardEncode7(%d)) = %d", msg, got)
		}
	}
}

func TestHadamard7CorrectsSingleErasureLikeZero(t *testing.T) {
	code := HadamardEncode7(93)
	code[17] = 0
	if got := HadamardDecode7(code[:]); got != 93 {
		t.Fatalf("HadamardDecode7 with one zero = %d, want 93", got)
	}
}
