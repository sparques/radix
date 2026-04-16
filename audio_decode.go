package radix

import (
	"fmt"
	"math"
)

type AlignedDecoderConfig struct {
	Audio AudioConfig
	Mode  Mode
}

func AnalyzeComplexAligned(cfg AlignedDecoderConfig, samples []complex64) (ToneFrames, error) {
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return nil, err
	}
	analyzer, err := newSymbolAnalyzer(cfg.Audio)
	if err != nil {
		return nil, err
	}

	frameCount := modeCfg.SymbolCount + 1
	wantLen := encodedSampleLen(modeCfg, analyzer.guardLen, analyzer.symbolLen)
	if len(samples) < wantLen {
		return nil, fmt.Errorf("got %d samples, want at least %d", len(samples), wantLen)
	}

	pos := 3 * (analyzer.guardLen + analyzer.symbolLen)
	frames := make(ToneFrames, frameCount)
	for i := 0; i < frameCount; i++ {
		symbol := samples[pos : pos+analyzer.symbolLen]
		frames[i] = analyzer.analyze(symbol)
		pos += analyzer.guardLen + analyzer.symbolLen
	}
	return frames, nil
}

func encodedSampleLen(cfg Config, guardLen, symbolLen int) int {
	return (cfg.SymbolCount + 4) * (guardLen + symbolLen)
}

type symbolAnalyzer struct {
	guardLen   int
	symbolLen  int
	toneOffset int
}

func newSymbolAnalyzer(cfg AudioConfig) (*symbolAnalyzer, error) {
	guardLen, err := cfg.GuardLen()
	if err != nil {
		return nil, err
	}
	symbolLen := guardLen * 40
	toneOffset, err := cfg.ToneOffset()
	if err != nil {
		return nil, err
	}
	return &symbolAnalyzer{
		guardLen:   guardLen,
		symbolLen:  symbolLen,
		toneOffset: toneOffset,
	}, nil
}

func (a *symbolAnalyzer) analyze(samples []complex64) []complex128 {
	scale := 0.5 / math.Sqrt(ToneCount)
	tones := make([]complex128, ToneCount)
	for i := range tones {
		carrier := bin(i+a.toneOffset, a.symbolLen)
		phase := -2 * math.Pi * float64(carrier) / float64(a.symbolLen)
		rot := complex(math.Cos(phase), math.Sin(phase))
		osc := complex(1, 0)
		var sum complex128
		for _, sample := range samples {
			sum += complex128(sample) * osc
			osc *= rot
		}
		tones[i] = sum / complex(scale*float64(a.symbolLen), 0)
	}
	return tones
}

func NearestSignedTone(v complex128) int8 {
	if real(v) < 0 {
		return -1
	}
	return 1
}
