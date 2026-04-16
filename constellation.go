package radix

import (
	"fmt"
	"math"
)

// Constellation maps signed bits to complex symbols and back.
// Think of a constellation as the little diagram of allowed points for BPSK,
// QPSK, or QAM. Bits are represented as -1/+1 values throughout this package.
type Constellation interface {
	// Modulation returns the modulation this mapper implements.
	Modulation() Modulation
	// Bits returns how many signed bits one symbol carries.
	Bits() int
	// Map converts signed bits into one complex constellation point.
	Map(bits []float64) (complex128, error)
	// Hard returns the nearest signed-bit decision for a received point.
	Hard(symbol complex128) []float64
	// Soft returns confidence-weighted signed bits for a received point.
	// Larger magnitudes mean stronger confidence.
	Soft(symbol complex128, precision float64) []float64
}

// NewConstellation returns the mapper/demapper for a modulation.
func NewConstellation(mod Modulation) (Constellation, error) {
	switch mod {
	case BPSK, QPSK, PSK8:
		return pskConstellation{mod: mod}, nil
	case QAM16, QAM64, QAM256, QAM1024, QAM4096:
		return qamConstellation{mod: mod}, nil
	default:
		return nil, fmt.Errorf("unsupported modulation %d", mod)
	}
}

type pskConstellation struct {
	mod Modulation
}

func (p pskConstellation) Modulation() Modulation {
	return p.mod
}

func (p pskConstellation) Bits() int {
	switch p.mod {
	case BPSK:
		return 1
	case QPSK:
		return 2
	case PSK8:
		return 3
	default:
		return 0
	}
}

func (p pskConstellation) Map(bits []float64) (complex128, error) {
	if err := requireBits(bits, p.Bits()); err != nil {
		return 0, err
	}

	switch p.mod {
	case BPSK:
		return complex(bits[0], 0), nil
	case QPSK:
		return complex(rcpSqrt2*bits[0], rcpSqrt2*bits[1]), nil
	case PSK8:
		real, imag := cosPi8, sinPi8
		if bits[0] < 0 {
			real, imag = imag, real
		}
		return complex(real*bits[1], imag*bits[2]), nil
	default:
		return 0, fmt.Errorf("unsupported PSK modulation %s", p.mod)
	}
}

func (p pskConstellation) Hard(symbol complex128) []float64 {
	bits := make([]float64, p.Bits())
	switch p.mod {
	case BPSK:
		bits[0] = sign(real(symbol))
	case QPSK:
		bits[0] = sign(real(symbol))
		bits[1] = sign(imag(symbol))
	case PSK8:
		bits[1] = sign(real(symbol))
		bits[2] = sign(imag(symbol))
		if math.Abs(real(symbol)) < math.Abs(imag(symbol)) {
			bits[0] = -1
		} else {
			bits[0] = 1
		}
	}
	return bits
}

func (p pskConstellation) Soft(symbol complex128, precision float64) []float64 {
	bits := make([]float64, p.Bits())
	switch p.mod {
	case BPSK:
		bits[0] = quantizeFloat(2, precision, real(symbol))
	case QPSK:
		dist := 2 * rcpSqrt2
		bits[0] = quantizeFloat(dist, precision, real(symbol))
		bits[1] = quantizeFloat(dist, precision, imag(symbol))
	case PSK8:
		dist := 2 * sinPi8
		bits[1] = quantizeFloat(dist, precision, real(symbol))
		bits[2] = quantizeFloat(dist, precision, imag(symbol))
		bits[0] = quantizeFloat(dist, precision, rcpSqrt2*(math.Abs(real(symbol))-math.Abs(imag(symbol))))
	}
	return bits
}

type qamConstellation struct {
	mod Modulation
}

func (q qamConstellation) Modulation() Modulation {
	return q.mod
}

func (q qamConstellation) Bits() int {
	switch q.mod {
	case QAM16:
		return 4
	case QAM64:
		return 6
	case QAM256:
		return 8
	case QAM1024:
		return 10
	case QAM4096:
		return 12
	default:
		return 0
	}
}

func (q qamConstellation) Map(bits []float64) (complex128, error) {
	if err := requireBits(bits, q.Bits()); err != nil {
		return 0, err
	}

	return complex(q.amp()*qamAxis(bits, 0), q.amp()*qamAxis(bits, 1)), nil
}

func (q qamConstellation) Hard(symbol complex128) []float64 {
	bits := make([]float64, q.Bits())
	amp := q.amp()
	hardQAMAxis(bits, 0, real(symbol), amp)
	hardQAMAxis(bits, 1, imag(symbol), amp)
	return bits
}

func (q qamConstellation) Soft(symbol complex128, precision float64) []float64 {
	bits := make([]float64, q.Bits())
	amp := q.amp()
	dist := 2 * amp
	softQAMAxis(bits, 0, real(symbol), amp, dist, precision)
	softQAMAxis(bits, 1, imag(symbol), amp, dist, precision)
	return bits
}

func (q qamConstellation) amp() float64 {
	num := 1 << q.Bits()
	return math.Sqrt(3.0 / (2.0 * float64(num-1)))
}

func qamAxis(bits []float64, offset int) float64 {
	axisBits := len(bits) / 2
	level := bits[offset+2*(axisBits-1)] + 2
	for i := axisBits - 2; i >= 1; i-- {
		step := 1 << (axisBits - i)
		level = bits[offset+2*i]*level + float64(step)
	}
	return bits[offset] * level
}

func hardQAMAxis(bits []float64, offset int, value, amp float64) {
	bits[offset] = sign(value)
	thresholdStep := 1 << (len(bits)/2 - 1)
	threshold := float64(thresholdStep)
	delta := math.Abs(value)
	for idx := offset + 2; idx < len(bits); idx += 2 {
		bits[idx] = sign(delta - amp*threshold)
		delta = math.Abs(delta - amp*threshold)
		threshold /= 2
	}
}

func softQAMAxis(bits []float64, offset int, value, amp, dist, precision float64) {
	bits[offset] = quantizeFloat(dist, precision, value)
	thresholdStep := 1 << (len(bits)/2 - 1)
	threshold := float64(thresholdStep)
	delta := math.Abs(value)
	for idx := offset + 2; idx < len(bits); idx += 2 {
		v := delta - amp*threshold
		bits[idx] = quantizeFloat(dist, precision, v)
		delta = math.Abs(v)
		threshold /= 2
	}
}

func requireBits(bits []float64, want int) error {
	if len(bits) != want {
		return fmt.Errorf("got %d bits, want %d", len(bits), want)
	}
	for idx, bit := range bits {
		if bit != -1 && bit != 1 {
			return fmt.Errorf("bit %d is %g, want -1 or +1", idx, bit)
		}
	}
	return nil
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func quantizeFloat(dist, precision, value float64) float64 {
	return value * dist * precision
}

const (
	rcpSqrt2 = 0.70710678118654752440
	cosPi8   = 0.92387953251128675613
	sinPi8   = 0.38268343236508977173
)
