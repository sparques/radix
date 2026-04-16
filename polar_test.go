package radix

import "testing"

func TestPolarEncodeRejectsWrongMessageLength(t *testing.T) {
	if _, err := PolarEncode(make([]int8, 71), Frozen256_72, 8); err == nil {
		t.Fatal("PolarEncode accepted short message")
	}
	if _, err := PolarEncode(make([]int8, 73), Frozen256_72, 8); err == nil {
		t.Fatal("PolarEncode accepted long message")
	}
}
