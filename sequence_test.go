package radix

import "testing"

func TestXorShift32Sequence(t *testing.T) {
	x := NewXorShift32(0)
	want := []uint32{723471715, 2497366906, 2064144800, 2008045182, 3532304609}
	for i, w := range want {
		if got := x.Next(); got != w {
			t.Fatalf("step %d = %d, want %d", i, got, w)
		}
	}
}

func TestXorShiftMaskSequence(t *testing.T) {
	x := NewXorShiftMask(8, 1, 1, 2, 1)
	want := []uint32{10, 85, 128, 192, 224, 240, 120, 252}
	for i, w := range want {
		if got := x.Next(); got != w {
			t.Fatalf("step %d = %d, want %d", i, got, w)
		}
	}
}

func TestMLSSequence(t *testing.T) {
	seq := NewMLS(MLS1Poly, 0)
	want := []bool{false, false, false, false, false, true, false, false}
	for i, w := range want {
		if got := seq.NextBit(); got != w {
			t.Fatalf("step %d = %t, want %t", i, got, w)
		}
	}
}
