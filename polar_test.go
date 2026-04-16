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

func TestPolarDecodeHardRoundTripMetadata(t *testing.T) {
	message := make([]int8, MetadataPolarBits)
	for i := range message {
		if i%3 == 0 {
			message[i] = -1
		} else {
			message[i] = 1
		}
	}
	code, err := PolarEncode(message, Frozen256_72, 8)
	if err != nil {
		t.Fatal(err)
	}
	got, err := PolarDecodeHard(code, Frozen256_72, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(message) {
		t.Fatalf("len = %d, want %d", len(got), len(message))
	}
	for i := range message {
		if got[i] != message[i] {
			t.Fatalf("message[%d] = %d, want %d", i, got[i], message[i])
		}
	}
}
