package radix

import (
	"fmt"
	"math"
	"sort"
)

type AlignedDecoderConfig struct {
	Audio AudioConfig
	Mode  Mode
}

type Acquisition struct {
	DataStart            int
	PreambleStart        int
	MatchedPreambleStart int
	Score                float64
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

	wantLen := encodedSampleLen(modeCfg, analyzer.guardLen, analyzer.symbolLen)
	if len(samples) < wantLen {
		return nil, fmt.Errorf("got %d samples, want at least %d", len(samples), wantLen)
	}

	pos := 3 * (analyzer.guardLen + analyzer.symbolLen)
	return analyzeComplexAt(modeCfg, analyzer, samples, pos)
}

func AnalyzeComplexAt(cfg AlignedDecoderConfig, samples []complex64, dataStart int) (ToneFrames, error) {
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return nil, err
	}
	analyzer, err := newSymbolAnalyzer(cfg.Audio)
	if err != nil {
		return nil, err
	}
	return analyzeComplexAt(modeCfg, analyzer, samples, dataStart)
}

func AcquireComplex(cfg AlignedDecoderConfig, samples []complex64) (Acquisition, error) {
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return Acquisition{}, err
	}
	analyzer, err := newSymbolAnalyzer(cfg.Audio)
	if err != nil {
		return Acquisition{}, err
	}
	candidates, err := acquireComplexCandidates(cfg.Audio, modeCfg, analyzer, samples)
	if err != nil {
		return Acquisition{}, err
	}
	return candidates[0], nil
}

func DecodeCaptured(cfg AlignedDecoderConfig, samples []complex64) (Metadata, []byte, Acquisition, error) {
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	analyzer, err := newSymbolAnalyzer(cfg.Audio)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	candidates, err := acquireComplexCandidates(cfg.Audio, modeCfg, analyzer, samples)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}

	var lastErr error
	for _, acquisition := range candidates {
		frames, err := analyzeComplexAt(modeCfg, analyzer, samples, acquisition.DataStart)
		if err != nil {
			lastErr = err
			continue
		}
		metadata, payload, err := decodeFramesForMode(cfg.Mode, modeCfg, frames)
		if err == nil {
			return metadata, payload, acquisition, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no acquisition candidate decoded")
	}
	return Metadata{}, nil, candidates[0], lastErr
}

func analyzeComplexAt(cfg Config, analyzer *symbolAnalyzer, samples []complex64, dataStart int) (ToneFrames, error) {
	if dataStart < 0 {
		return nil, fmt.Errorf("data start %d is before start of samples", dataStart)
	}
	frameCount := cfg.SymbolCount + 1
	stride := analyzer.guardLen + analyzer.symbolLen
	wantLen := dataStart + (frameCount-1)*stride + analyzer.symbolLen
	if len(samples) < wantLen {
		return nil, fmt.Errorf("got %d samples, want at least %d for data start %d", len(samples), wantLen, dataStart)
	}

	pos := dataStart
	frames := make(ToneFrames, frameCount)
	for i := 0; i < frameCount; i++ {
		symbol := samples[pos : pos+analyzer.symbolLen]
		frames[i] = analyzer.analyze(symbol)
		pos += stride
	}
	return frames, nil
}

func acquireComplexCandidates(audio AudioConfig, modeCfg Config, analyzer *symbolAnalyzer, samples []complex64) ([]Acquisition, error) {
	template, err := schmidlCoxTemplate(audio)
	if err != nil {
		return nil, err
	}
	if len(samples) < len(template) {
		return nil, fmt.Errorf("got %d samples, want at least %d", len(samples), len(template))
	}

	step := analyzer.guardLen / 8
	if step < 1 {
		step = 1
	}
	top := make([]bodyCandidate, 0, 12)
	for pos := 0; pos+len(template) <= len(samples); pos += step {
		top = insertBodyCandidate(top, bodyCandidate{start: pos, score: correlationScore(samples[pos:pos+len(template)], template)}, 12)
	}

	refined := make([]bodyCandidate, 0, len(top))
	seenBody := map[int]bool{}
	for _, candidate := range top {
		start := candidate.start - step + 1
		if start < 0 {
			start = 0
		}
		stop := candidate.start + step - 1
		maxStart := len(samples) - len(template)
		if stop > maxStart {
			stop = maxStart
		}
		best := candidate
		for pos := start; pos <= stop; pos++ {
			score := correlationScore(samples[pos:pos+len(template)], template)
			if score > best.score {
				best = bodyCandidate{start: pos, score: score}
			}
		}
		if !seenBody[best.start] {
			refined = append(refined, best)
			seenBody[best.start] = true
		}
	}
	sort.Slice(refined, func(i, j int) bool {
		return refined[i].score > refined[j].score
	})

	frameCount := modeCfg.SymbolCount + 1
	stride := analyzer.guardLen + analyzer.symbolLen
	seenData := map[int]bool{}
	acquisitions := make([]Acquisition, 0, 2*len(refined))
	for _, candidate := range refined {
		hypotheses := []struct {
			dataStart int
			pairStart int
		}{
			{
				dataStart: candidate.start + analyzer.guardLen + analyzer.symbolLen,
				pairStart: candidate.start - stride,
			},
			{
				dataStart: candidate.start + analyzer.guardLen + 2*analyzer.symbolLen,
				pairStart: candidate.start + stride,
			},
		}
		for _, hypothesis := range hypotheses {
			dataStart := hypothesis.dataStart
			required := dataStart + (frameCount-1)*stride + analyzer.symbolLen
			if required > len(samples) || seenData[dataStart] {
				continue
			}
			pairScore := 0.0
			if hypothesis.pairStart >= 0 && hypothesis.pairStart+len(template) <= len(samples) {
				pairScore = correlationScore(samples[hypothesis.pairStart:hypothesis.pairStart+len(template)], template)
			}
			score := math.Min(candidate.score, pairScore)
			acquisitions = append(acquisitions, Acquisition{
				DataStart:            dataStart,
				PreambleStart:        dataStart - 3*stride,
				MatchedPreambleStart: candidate.start,
				Score:                score,
			})
			seenData[dataStart] = true
		}
	}
	if len(acquisitions) == 0 {
		return nil, fmt.Errorf("no complete captured frame found")
	}
	sort.Slice(acquisitions, func(i, j int) bool {
		if acquisitions[i].Score == acquisitions[j].Score {
			return acquisitions[i].DataStart < acquisitions[j].DataStart
		}
		return acquisitions[i].Score > acquisitions[j].Score
	})
	return acquisitions, nil
}

type bodyCandidate struct {
	start int
	score float64
}

func insertBodyCandidate(candidates []bodyCandidate, candidate bodyCandidate, limit int) []bodyCandidate {
	candidates = append(candidates, candidate)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func schmidlCoxTemplate(cfg AudioConfig) ([]complex64, error) {
	synth, err := newSymbolSynthesizer(cfg)
	if err != nil {
		return nil, err
	}
	seq := NewMLS(MLS0Poly, MLS0Seed)
	tone := make([]complex128, ToneCount)
	for i := range tone {
		tone[i] = complex(float64(NRZ(seq.NextBit())), 0)
	}
	tdom := clipComplex(synth.synthesize(tone))
	out := make([]complex64, len(tdom))
	for i, sample := range tdom {
		out[i] = complex64(sample)
	}
	return out, nil
}

func correlationScore(samples, template []complex64) float64 {
	var cross complex128
	var sampleEnergy, templateEnergy float64
	for i, sample := range samples {
		s := complex128(sample)
		t := complex128(template[i])
		cross += s * complex(real(t), -imag(t))
		sampleEnergy += real(s)*real(s) + imag(s)*imag(s)
		templateEnergy += real(t)*real(t) + imag(t)*imag(t)
	}
	if sampleEnergy == 0 || templateEnergy == 0 {
		return 0
	}
	return math.Hypot(real(cross), imag(cross)) / math.Sqrt(sampleEnergy*templateEnergy)
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
