package radix

import "fmt"

const Bandwidth = 2400

type ToneKind uint8

const (
	SeedTone ToneKind = iota
	MetaTone
	DataTone
)

func (k ToneKind) String() string {
	switch k {
	case SeedTone:
		return "seed"
	case MetaTone:
		return "meta"
	case DataTone:
		return "data"
	default:
		return fmt.Sprintf("ToneKind(%d)", k)
	}
}

type ToneSlot struct {
	Index     int
	Kind      ToneKind
	Bits      int
	BitOffset int
}

type SymbolPlan struct {
	Index      int
	SeedOffset int
	Tones      []ToneSlot
}

func ValidateFrequencyOffset(sampleRate, channels, offset int) error {
	if offset%300 != 0 {
		return fmt.Errorf("frequency offset %d is not divisible by 300", offset)
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("unsupported channel count %d", channels)
	}
	if sampleRate != 44100 && sampleRate != 48000 {
		return fmt.Errorf("unsupported sample rate %d", sampleRate)
	}
	if (channels == 1 && offset < Bandwidth/2) || offset < Bandwidth/2-sampleRate/2 || offset > sampleRate/2-Bandwidth/2 {
		return fmt.Errorf("unsupported frequency offset %d for %d Hz and %d channel(s)", offset, sampleRate, channels)
	}
	return nil
}

func ToneOffset(sampleRate, offset int) (int, error) {
	if sampleRate != 44100 && sampleRate != 48000 {
		return 0, fmt.Errorf("unsupported sample rate %d", sampleRate)
	}
	guardLen := sampleRate / 300
	symbolLen := guardLen * 40
	return (offset*symbolLen)/sampleRate - ToneCount/2, nil
}

func BuildTonePlan(cfg Config) ([]SymbolPlan, error) {
	if cfg.SymbolCount <= 0 {
		return nil, fmt.Errorf("invalid symbol count %d", cfg.SymbolCount)
	}
	if cfg.ModBits <= 0 {
		return nil, fmt.Errorf("invalid modulation bits %d", cfg.ModBits)
	}

	plans := make([]SymbolPlan, cfg.SymbolCount+1)
	dataOffset := 0
	metaOffset := 0
	for j := range plans {
		seedOffset := (BlockSkew*j + FirstSeed) % BlockLength
		plan := SymbolPlan{
			Index:      j,
			SeedOffset: seedOffset,
			Tones:      make([]ToneSlot, 0, ToneCount),
		}
		for i := 0; i < ToneCount; i++ {
			slot := ToneSlot{Index: i}
			if i%BlockLength == seedOffset {
				slot.Kind = SeedTone
			} else if j == 0 {
				slot.Kind = MetaTone
				slot.Bits = 1
				slot.BitOffset = metaOffset
				metaOffset++
			} else {
				bits := dataToneBits(cfg.ModBits, dataOffset)
				slot.Kind = DataTone
				slot.Bits = bits
				slot.BitOffset = dataOffset
				dataOffset += bits
			}
			plan.Tones = append(plan.Tones, slot)
		}
		plans[j] = plan
	}

	codeBits := 1 << cfg.CodeOrder
	if metaOffset != DataTones {
		return nil, fmt.Errorf("metadata plan uses %d bits, want %d", metaOffset, DataTones)
	}
	if dataOffset != codeBits {
		return nil, fmt.Errorf("data plan uses %d bits, want %d", dataOffset, codeBits)
	}
	return plans, nil
}

func dataToneBits(modBits, bitOffset int) int {
	switch {
	case modBits == 3 && bitOffset%32 == 30:
		return 2
	case modBits == 6 && bitOffset%64 == 60:
		return 4
	case modBits == 10 && bitOffset%128 == 120:
		return 8
	case modBits == 12 && bitOffset%128 == 120:
		return 8
	default:
		return modBits
	}
}
