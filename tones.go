package radix

import "fmt"

type ToneFrames [][]complex128

func BuildToneFrames(cfg Config, metadata []int8, payload []int8) (ToneFrames, error) {
	if len(metadata) != MetadataCodeBits {
		return nil, fmt.Errorf("metadata has %d symbols, want %d", len(metadata), MetadataCodeBits)
	}
	codeBits := 1 << cfg.CodeOrder
	if len(payload) != codeBits {
		return nil, fmt.Errorf("payload has %d symbols, want %d", len(payload), codeBits)
	}

	plan, err := BuildTonePlan(cfg)
	if err != nil {
		return nil, err
	}
	frames := make(ToneFrames, len(plan))
	seed := NewMLS(MLS1Poly, 0)
	constellations := map[int]Constellation{}
	for _, bits := range []int{1, 2, 3, 4, 6, 8, 10, 12} {
		mod, err := modulationForBits(bits)
		if err != nil {
			return nil, err
		}
		c, err := NewConstellation(mod)
		if err != nil {
			return nil, err
		}
		constellations[bits] = c
	}

	for j, symbolPlan := range plan {
		frame := make([]complex128, ToneCount)
		for _, tone := range symbolPlan.Tones {
			switch tone.Kind {
			case SeedTone:
				frame[tone.Index] = complex(float64(NRZ(seed.NextBit())), 0)
			case MetaTone:
				frame[tone.Index] = complex(float64(metadata[tone.BitOffset]), 0)
			case DataTone:
				bits := int8BitsToFloat64(payload[tone.BitOffset : tone.BitOffset+tone.Bits])
				symbol, err := constellations[tone.Bits].Map(bits)
				if err != nil {
					return nil, err
				}
				frame[tone.Index] = symbol
			default:
				return nil, fmt.Errorf("unsupported tone kind %s", tone.Kind)
			}
		}
		frames[j] = frame
	}
	return frames, nil
}

func DecodeToneFrames(cfg Config, frames ToneFrames) ([]int8, []int8, error) {
	plan, err := BuildTonePlan(cfg)
	if err != nil {
		return nil, nil, err
	}
	if len(frames) != len(plan) {
		return nil, nil, fmt.Errorf("got %d tone frames, want %d", len(frames), len(plan))
	}

	metadata := make([]int8, MetadataCodeBits)
	payload := make([]int8, 1<<cfg.CodeOrder)
	constellations := map[int]Constellation{}
	for _, bits := range []int{1, 2, 3, 4, 6, 8, 10, 12} {
		mod, err := modulationForBits(bits)
		if err != nil {
			return nil, nil, err
		}
		c, err := NewConstellation(mod)
		if err != nil {
			return nil, nil, err
		}
		constellations[bits] = c
	}

	for frameIdx, symbolPlan := range plan {
		if len(frames[frameIdx]) != ToneCount {
			return nil, nil, fmt.Errorf("frame %d has %d tones, want %d", frameIdx, len(frames[frameIdx]), ToneCount)
		}
		for _, tone := range symbolPlan.Tones {
			symbol := frames[frameIdx][tone.Index]
			switch tone.Kind {
			case SeedTone:
				continue
			case MetaTone:
				metadata[tone.BitOffset] = NearestSignedTone(symbol)
			case DataTone:
				bits := constellations[tone.Bits].Hard(symbol)
				for i, bit := range bits {
					payload[tone.BitOffset+i] = int8(bit)
				}
			default:
				return nil, nil, fmt.Errorf("unsupported tone kind %s", tone.Kind)
			}
		}
	}
	return metadata, payload, nil
}

func DecodeAligned(cfg AlignedDecoderConfig, samples []complex64) (Metadata, []byte, error) {
	frames, err := AnalyzeComplexAligned(cfg, samples)
	if err != nil {
		return Metadata{}, nil, err
	}
	modeCfg, err := Setup(cfg.Mode)
	if err != nil {
		return Metadata{}, nil, err
	}
	return decodeFramesForMode(cfg.Mode, modeCfg, frames)
}

func decodeFramesForMode(mode Mode, cfg Config, frames ToneFrames) (Metadata, []byte, error) {
	metadataCode, payloadCode, err := DecodeToneFrames(cfg, frames)
	if err != nil {
		return Metadata{}, nil, err
	}
	metadata, err := DecodeMetadata(metadataCode)
	if err != nil {
		return Metadata{}, nil, err
	}
	if metadata.Mode != mode {
		return Metadata{}, nil, fmt.Errorf("metadata mode %d does not match decoder mode %d", metadata.Mode, mode)
	}
	payload, err := DecodePayload(cfg, payloadCode)
	if err != nil {
		return Metadata{}, nil, err
	}
	return metadata, payload, nil
}

func int8BitsToFloat64(bits []int8) []float64 {
	out := make([]float64, len(bits))
	for i, bit := range bits {
		out[i] = float64(bit)
	}
	return out
}

func modulationForBits(bits int) (Modulation, error) {
	switch bits {
	case 1:
		return BPSK, nil
	case 2:
		return QPSK, nil
	case 3:
		return PSK8, nil
	case 4:
		return QAM16, nil
	case 6:
		return QAM64, nil
	case 8:
		return QAM256, nil
	case 10:
		return QAM1024, nil
	case 12:
		return QAM4096, nil
	default:
		return 0, fmt.Errorf("unsupported modulation bit count %d", bits)
	}
}
