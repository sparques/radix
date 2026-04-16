package radix

import (
	"fmt"
	"math"
)

type AudioConfig struct {
	SampleRate      int
	FrequencyOffset int
}

func (c AudioConfig) GuardLen() (int, error) {
	if c.SampleRate != 44100 && c.SampleRate != 48000 {
		return 0, fmt.Errorf("unsupported sample rate %d", c.SampleRate)
	}
	return c.SampleRate / 300, nil
}

func (c AudioConfig) SymbolLen() (int, error) {
	guardLen, err := c.GuardLen()
	if err != nil {
		return 0, err
	}
	return guardLen * 40, nil
}

func (c AudioConfig) ToneOffset() (int, error) {
	return ToneOffset(c.SampleRate, c.FrequencyOffset)
}

type EncoderConfig struct {
	Audio    AudioConfig
	Mode     Mode
	CallSign int64
}

func EncodeComplex(cfg EncoderConfig, payload []byte) ([]complex64, error) {
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return nil, err
	}
	if cfg.CallSign <= 0 || cfg.CallSign >= MaxCallSign {
		return nil, fmt.Errorf("unsupported call sign value %d", cfg.CallSign)
	}
	if _, err := cfg.Audio.SymbolLen(); err != nil {
		return nil, err
	}

	meta, err := EncodeMetadata(cfg.CallSign, cfg.Mode)
	if err != nil {
		return nil, err
	}
	payloadCode, err := EncodePayload(modeCfg, payload)
	if err != nil {
		return nil, err
	}
	frames, err := BuildToneFrames(modeCfg, meta, payloadCode)
	if err != nil {
		return nil, err
	}

	var out []complex64
	state, err := newSymbolSynthesizer(cfg.Audio)
	if err != nil {
		return nil, err
	}
	out = append(out, state.leadingNoise()...)
	out = append(out, state.schmidlCox()...)
	for j, frame := range frames {
		out = append(out, state.symbol(frame, j)...)
	}
	out = append(out, state.finish()...)
	return out, nil
}

func ComplexToInterleavedFloat32(samples []complex64) []float32 {
	out := make([]float32, 2*len(samples))
	for i, sample := range samples {
		out[2*i] = real(sample)
		out[2*i+1] = imag(sample)
	}
	return out
}

func ComplexToMonoFloat32(samples []complex64) []float32 {
	out := make([]float32, len(samples))
	for i, sample := range samples {
		out[i] = real(sample)
	}
	return out
}

type symbolSynthesizer struct {
	sampleRate int
	guardLen   int
	symbolLen  int
	toneOffset int
	weight     []float64
	guard      []complex128
}

func newSymbolSynthesizer(cfg AudioConfig) (*symbolSynthesizer, error) {
	guardLen, err := cfg.GuardLen()
	if err != nil {
		return nil, err
	}
	symbolLen := guardLen * 40
	toneOffset, err := cfg.ToneOffset()
	if err != nil {
		return nil, err
	}
	return &symbolSynthesizer{
		sampleRate: cfg.SampleRate,
		guardLen:   guardLen,
		symbolLen:  symbolLen,
		toneOffset: toneOffset,
		weight:     guardWeights(guardLen),
		guard:      make([]complex128, guardLen),
	}, nil
}

func (s *symbolSynthesizer) leadingNoise() []complex64 {
	noise := NewMLS(MLS2Poly, 0)
	tone := make([]complex128, ToneCount)
	for i := range tone {
		tone[i] = complex(float64(NRZ(noise.NextBit())), 0)
	}
	return s.symbol(tone, -3)
}

func (s *symbolSynthesizer) schmidlCox() []complex64 {
	seq := NewMLS(MLS0Poly, MLS0Seed)
	tone := make([]complex128, ToneCount)
	for i := range tone {
		tone[i] = complex(float64(NRZ(seq.NextBit())), 0)
	}
	out := s.symbol(tone, -2)
	out = append(out, s.symbol(tone, -1)...)
	return out
}

func (s *symbolSynthesizer) symbol(tone []complex128, symbolNumber int) []complex64 {
	tdom := s.synthesize(tone)

	if symbolNumber >= 0 {
		tdom = applySeedZero(tone, tdom, s)
	}
	tdom = clipComplex(tdom)

	out := make([]complex64, 0, s.guardLen+s.symbolLen)
	if symbolNumber != -1 {
		for i := 0; i < s.guardLen; i++ {
			g := lerpComplex(s.guard[i], tdom[i+s.symbolLen-s.guardLen], s.weight[i])
			out = append(out, complex64(g))
		}
	}
	for i := 0; i < s.guardLen; i++ {
		s.guard[i] = tdom[i]
	}
	for _, sample := range tdom {
		out = append(out, complex64(sample))
	}
	return out
}

func (s *symbolSynthesizer) finish() []complex64 {
	out := make([]complex64, s.guardLen)
	for i := 0; i < s.guardLen; i++ {
		out[i] = complex64(s.guard[i] * complex(1-s.weight[i], 0))
		s.guard[i] = 0
	}
	return out
}

func (s *symbolSynthesizer) synthesize(tone []complex128) []complex128 {
	scale := 0.5 / math.Sqrt(ToneCount)
	out := make([]complex128, s.symbolLen)
	for i, t := range tone {
		carrier := bin(i+s.toneOffset, s.symbolLen)
		phase := 2 * math.Pi * float64(carrier) / float64(s.symbolLen)
		rot := complex(math.Cos(phase), math.Sin(phase))
		osc := complex(1, 0)
		for n := range out {
			out[n] += t * osc
			osc *= rot
		}
	}
	for n := range out {
		out[n] *= complex(scale, 0)
	}
	return out
}

func applySeedZero(_ []complex128, tdom []complex128, _ *symbolSynthesizer) []complex128 {
	// Upstream searches all 128 Hadamard seed words to reduce PAPR. Seed 0 is
	// protocol-valid and applies no MLS2 scrambling, so this preserves a
	// decodable signal while leaving PAPR optimization for a later pass.
	return tdom
}

func guardWeights(guardLen int) []float64 {
	weight := make([]float64, guardLen)
	for i := 0; i < guardLen/4; i++ {
		weight[i] = 0
	}
	start := guardLen / 4
	stop := start + guardLen/2
	for i := start; i < stop; i++ {
		x := float64(i-start) / float64(guardLen/2-1)
		weight[i] = 0.5 * (1 - math.Cos(math.Pi*x))
	}
	for i := stop; i < guardLen; i++ {
		weight[i] = 1
	}
	return weight
}

func clipComplex(samples []complex128) []complex128 {
	for i, sample := range samples {
		power := real(sample)*real(sample) + imag(sample)*imag(sample)
		if power > 1 {
			samples[i] = sample / complex(math.Sqrt(power), 0)
		}
		samples[i] = complex(clamp(real(samples[i]), -1, 1), clamp(imag(samples[i]), -1, 1))
	}
	return samples
}

func bin(carrier, symbolLen int) int {
	return (carrier + symbolLen) % symbolLen
}

func lerpComplex(a, b complex128, weight float64) complex128 {
	return a + complex(weight, 0)*(b-a)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
