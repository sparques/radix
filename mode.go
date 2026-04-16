package radix

import "fmt"

const (
	ModMax     = 12
	CodeMax    = 16
	BitsMax    = 1 << CodeMax
	DataMax    = 8192
	SymbolsMax = 26 + 1

	MLS0Poly = 0x331
	MLS0Seed = 214
	MLS1Poly = 0x43
	MLS2Poly = 0x163

	DataTones   = 256
	SeedTones   = 64
	ToneCount   = DataTones + SeedTones
	BlockLength = 5
	BlockSkew   = 3
	FirstSeed   = 4
)

type Modulation uint8

const (
	BPSK Modulation = iota
	QPSK
	PSK8
	QAM16
	QAM64
	QAM256
	QAM1024
	QAM4096
)

func (m Modulation) String() string {
	switch m {
	case BPSK:
		return "BPSK"
	case QPSK:
		return "QPSK"
	case PSK8:
		return "8PSK"
	case QAM16:
		return "QAM16"
	case QAM64:
		return "QAM64"
	case QAM256:
		return "QAM256"
	case QAM1024:
		return "QAM1024"
	case QAM4096:
		return "QAM4096"
	default:
		return fmt.Sprintf("Modulation(%d)", m)
	}
}

type CodeRate uint8

const (
	RateHalf CodeRate = iota
	RateTwoThirds
	RateThreeQuarters
	RateFiveSixths
)

func (r CodeRate) String() string {
	switch r {
	case RateHalf:
		return "1/2"
	case RateTwoThirds:
		return "2/3"
	case RateThreeQuarters:
		return "3/4"
	case RateFiveSixths:
		return "5/6"
	default:
		return fmt.Sprintf("CodeRate(%d)", r)
	}
}

type FrameSize uint8

const (
	ShortFrame FrameSize = iota
	NormalFrame
)

func (s FrameSize) String() string {
	switch s {
	case ShortFrame:
		return "short"
	case NormalFrame:
		return "normal"
	default:
		return fmt.Sprintf("FrameSize(%d)", s)
	}
}

type Mode uint8

func NewMode(mod Modulation, rate CodeRate, frame FrameSize) (Mode, error) {
	if mod > QAM4096 {
		return 0, fmt.Errorf("unsupported modulation %d", mod)
	}
	if rate > RateFiveSixths {
		return 0, fmt.Errorf("unsupported code rate %d", rate)
	}
	if frame > NormalFrame {
		return 0, fmt.Errorf("unsupported frame size %d", frame)
	}
	return Mode(uint8(mod)<<4 | uint8(rate)<<1 | uint8(frame)), nil
}

func (m Mode) Modulation() Modulation {
	return Modulation((m >> 4) & 7)
}

func (m Mode) CodeRate() CodeRate {
	return CodeRate((m >> 1) & 7)
}

func (m Mode) FrameSize() FrameSize {
	return FrameSize(m & 1)
}

func (m Mode) Analog() bool {
	return m&128 != 0
}

type Config struct {
	Mode        Mode
	Modulation  Modulation
	CodeRate    CodeRate
	FrameSize   FrameSize
	ModBits     int
	DataBits    int
	DataBytes   int
	CodeOrder   int
	SymbolCount int
	Duration    float64
	BitrateKbps float64
}

func Setup(mode Mode) (Config, error) {
	if mode.Analog() {
		return Config{}, fmt.Errorf("analog mode is not supported")
	}

	cfg := Config{
		Mode:       mode,
		Modulation: mode.Modulation(),
		CodeRate:   mode.CodeRate(),
		FrameSize:  mode.FrameSize(),
	}

	switch cfg.Modulation {
	case BPSK:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 1, 8, 11
	case QPSK:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 2, 4, 11
	case PSK8:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 3, 11, 13
	case QAM16:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 4, 4, 12
	case QAM64:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 6, 11, 14
	case QAM256:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 8, 8, 14
	case QAM1024:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 10, 13, 15
	case QAM4096:
		cfg.ModBits, cfg.SymbolCount, cfg.CodeOrder = 12, 11, 15
	default:
		return Config{}, fmt.Errorf("unsupported modulation %d", cfg.Modulation)
	}

	if cfg.FrameSize == NormalFrame {
		if cfg.SymbolCount == 4 {
			cfg.SymbolCount *= 4
			cfg.CodeOrder += 2
		} else {
			cfg.SymbolCount *= 2
			cfg.CodeOrder++
		}
	} else if cfg.FrameSize != ShortFrame {
		return Config{}, fmt.Errorf("unsupported frame size %d", cfg.FrameSize)
	}

	dataBits, err := dataBits(cfg.CodeOrder, cfg.CodeRate)
	if err != nil {
		return Config{}, err
	}
	cfg.DataBits = dataBits
	cfg.DataBytes = dataBits / 8
	cfg.Duration = 41.0 / 300.0 * float64(3+cfg.SymbolCount)
	cfg.BitrateKbps = float64(cfg.DataBits) / cfg.Duration / 1000.0

	return cfg, nil
}

func dataBits(codeOrder int, rate CodeRate) (int, error) {
	if codeOrder < 11 || codeOrder > 16 {
		return 0, fmt.Errorf("unsupported code order %d", codeOrder)
	}

	table := map[CodeRate][6]int{
		RateHalf:          {1024, 2048, 4096, 8192, 16384, 32768},
		RateTwoThirds:     {1368, 2736, 5472, 10944, 21888, 43776},
		RateThreeQuarters: {1536, 3072, 6144, 12288, 24576, 49152},
		RateFiveSixths:    {1704, 3408, 6816, 13632, 27264, 54528},
	}
	bits, ok := table[rate]
	if !ok {
		return 0, fmt.Errorf("unsupported code rate %d", rate)
	}
	return bits[codeOrder-11], nil
}
