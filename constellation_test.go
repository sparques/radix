package radix

import (
	"math/cmplx"
	"testing"
)

func TestConstellationMapHardRoundTrip(t *testing.T) {
	for _, mod := range []Modulation{BPSK, QPSK, PSK8, QAM16, QAM64, QAM256, QAM1024, QAM4096} {
		c, err := NewConstellation(mod)
		if err != nil {
			t.Fatalf("NewConstellation(%s): %v", mod, err)
		}
		for _, bits := range allSignedBits(c.Bits()) {
			symbol, err := c.Map(bits)
			if err != nil {
				t.Fatalf("%s Map(%v): %v", mod, bits, err)
			}
			got := c.Hard(symbol)
			for idx := range bits {
				if got[idx] != bits[idx] {
					t.Fatalf("%s Hard(Map(%v)) = %v", mod, bits, got)
				}
			}
		}
	}
}

func TestKnownConstellationPoints(t *testing.T) {
	tests := []struct {
		mod  Modulation
		bits []float64
		want complex128
	}{
		{BPSK, []float64{1}, complex(1, 0)},
		{QPSK, []float64{1, -1}, complex(rcpSqrt2, -rcpSqrt2)},
		{PSK8, []float64{-1, 1, -1}, complex(sinPi8, -cosPi8)},
		{QAM16, []float64{1, -1, 1, -1}, complex(3*qamAmp(QAM16), -qamAmp(QAM16))},
		{QAM64, []float64{-1, 1, -1, 1, 1, -1}, complex(-qamAmp(QAM64), 5*qamAmp(QAM64))},
	}

	for _, tt := range tests {
		c, err := NewConstellation(tt.mod)
		if err != nil {
			t.Fatalf("NewConstellation(%s): %v", tt.mod, err)
		}
		got, err := c.Map(tt.bits)
		if err != nil {
			t.Fatalf("%s Map(%v): %v", tt.mod, tt.bits, err)
		}
		if cmplx.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s Map(%v) = %v, want %v", tt.mod, tt.bits, got, tt.want)
		}
	}
}

func TestConstellationRejectsInvalidBits(t *testing.T) {
	c, err := NewConstellation(QPSK)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Map([]float64{1}); err == nil {
		t.Fatal("Map accepted wrong bit count")
	}
	if _, err := c.Map([]float64{0, 1}); err == nil {
		t.Fatal("Map accepted bit outside signed convention")
	}
}

func allSignedBits(count int) [][]float64 {
	out := make([][]float64, 1<<count)
	for mask := range out {
		bits := make([]float64, count)
		for bit := range bits {
			if mask&(1<<bit) == 0 {
				bits[bit] = -1
			} else {
				bits[bit] = 1
			}
		}
		out[mask] = bits
	}
	return out
}

func qamAmp(mod Modulation) float64 {
	c, err := NewConstellation(mod)
	if err != nil {
		panic(err)
	}
	return c.(qamConstellation).amp()
}
