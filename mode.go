package radix

import "fmt"

const (
	// ModMax is the largest number of bits any supported modulation can place on
	// one data tone. Most callers use Modulation values instead.
	ModMax = 12
	// CodeMax is the largest polar-code order supported by the mode tables.
	CodeMax = 16
	// BitsMax is the largest encoded payload bit count supported by the package.
	BitsMax = 1 << CodeMax
	// DataMax is the largest uncoded payload size, in bytes, for any mode.
	DataMax = 8192
	// SymbolsMax is the maximum number of OFDM data symbols plus metadata.
	SymbolsMax = 26 + 1

	// MLS0Poly is the maximal-length-sequence polynomial used by the sync
	// preamble. You normally do not need this unless you are inspecting the wire
	// format.
	MLS0Poly = 0x331
	// MLS0Seed is the starting state for the sync preamble sequence.
	MLS0Seed = 214
	// MLS1Poly is the maximal-length-sequence polynomial used for seed tones.
	MLS1Poly = 0x43
	// MLS2Poly is the maximal-length-sequence polynomial used by the transmitter
	// noise and PAPR-related parts of the format.
	MLS2Poly = 0x163

	// DataTones is the number of OFDM tones that carry metadata or payload in a
	// symbol.
	DataTones = 256
	// SeedTones is the number of OFDM tones reserved as known reference tones.
	// The receiver uses them to correct phase and gain.
	SeedTones = 64
	// ToneCount is the total number of active OFDM tones in one symbol.
	ToneCount = DataTones + SeedTones
	// BlockLength is the spacing pattern used to interleave seed tones among data
	// tones.
	BlockLength = 5
	// BlockSkew is the per-symbol shift of the seed-tone pattern.
	BlockSkew = 3
	// FirstSeed is the first seed-tone slot in the tone pattern.
	FirstSeed = 4
)

// Modulation selects how many possible symbols are sent on each data tone.
// BPSK is easiest to receive but slowest; larger QAM modes are faster and need
// a cleaner channel.
type Modulation uint8

const (
	// BPSK sends one bit per data tone and is the most robust modulation.
	BPSK Modulation = iota
	// QPSK sends two bits per data tone.
	QPSK
	// PSK8 sends three bits per data tone.
	PSK8
	// QAM16 sends four bits per data tone.
	QAM16
	// QAM64 sends six bits per data tone.
	QAM64
	// QAM256 sends eight bits per data tone.
	QAM256
	// QAM1024 sends ten bits per data tone.
	QAM1024
	// QAM4096 sends twelve bits per data tone and needs the cleanest channel.
	QAM4096
)

// String returns the protocol spelling for the modulation, such as "QPSK" or
// "QAM16".
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

// ParseModulation parses the strings produced by Modulation.String.
func ParseModulation(s string) (Modulation, error) {
	switch s {
	case "BPSK":
		return BPSK, nil
	case "QPSK":
		return QPSK, nil
	case "8PSK":
		return PSK8, nil
	case "QAM16":
		return QAM16, nil
	case "QAM64":
		return QAM64, nil
	case "QAM256":
		return QAM256, nil
	case "QAM1024":
		return QAM1024, nil
	case "QAM4096":
		return QAM4096, nil
	default:
		return 0, fmt.Errorf("unsupported modulation %q", s)
	}
}

// CodeRate selects how much forward-error-correction redundancy is added.
// Lower rates are slower but leave more room for the decoder to recover from
// bit errors.
type CodeRate uint8

const (
	// RateHalf is the most redundant supported code rate.
	RateHalf CodeRate = iota
	// RateTwoThirds carries more data than RateHalf with less redundancy.
	RateTwoThirds
	// RateThreeQuarters is a middle/high throughput code rate.
	RateThreeQuarters
	// RateFiveSixths is the least redundant supported code rate.
	RateFiveSixths
)

// String returns the protocol spelling for the code rate, such as "1/2".
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

// ParseCodeRate parses the strings produced by CodeRate.String.
func ParseCodeRate(s string) (CodeRate, error) {
	switch s {
	case "1/2":
		return RateHalf, nil
	case "2/3":
		return RateTwoThirds, nil
	case "3/4":
		return RateThreeQuarters, nil
	case "5/6":
		return RateFiveSixths, nil
	default:
		return 0, fmt.Errorf("unsupported code rate %q", s)
	}
}

// FrameSize selects the short or normal form of the on-air frame.
// Short frames finish sooner; normal frames carry more data per transmission.
type FrameSize uint8

const (
	// ShortFrame selects the shorter frame format.
	ShortFrame FrameSize = iota
	// NormalFrame selects the normal, higher-capacity frame format.
	NormalFrame
)

// String returns "short" or "normal".
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

// ParseFrameSize parses the strings produced by FrameSize.String.
func ParseFrameSize(s string) (FrameSize, error) {
	switch s {
	case "short":
		return ShortFrame, nil
	case "normal":
		return NormalFrame, nil
	default:
		return 0, fmt.Errorf("unsupported frame size %q", s)
	}
}

// Mode is the compact protocol byte that combines modulation, code rate, and
// frame size. Use NewMode to build one instead of packing bits by hand.
type Mode uint8

// NewMode packs modulation, code rate, and frame size into a Mode after checking
// that each part is supported.
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

// Modulation extracts the modulation part of a Mode.
func (m Mode) Modulation() Modulation {
	return Modulation((m >> 4) & 7)
}

// CodeRate extracts the forward-error-correction rate part of a Mode.
func (m Mode) CodeRate() CodeRate {
	return CodeRate((m >> 1) & 7)
}

// FrameSize extracts the short/normal frame selection part of a Mode.
func (m Mode) FrameSize() FrameSize {
	return FrameSize(m & 1)
}

// Analog reports whether the analog-mode bit is set. Radix currently implements
// the digital modem path, so Setup rejects analog modes.
func (m Mode) Analog() bool {
	return m&128 != 0
}

// Config is the expanded, easy-to-use description of a Mode.
// It includes derived sizes such as how many payload bytes fit and how many OFDM
// symbols the frame uses.
type Config struct {
	// Mode is the compact protocol mode byte this config was derived from.
	Mode Mode
	// Modulation is the modulation selected by Mode.
	Modulation Modulation
	// CodeRate is the forward-error-correction rate selected by Mode.
	CodeRate CodeRate
	// FrameSize is the short/normal frame setting selected by Mode.
	FrameSize FrameSize
	// FrozenBits is the polar-code frozen-bit table for this payload mode.
	FrozenBits []uint32
	// ModBits is the number of bits carried by a full data tone.
	ModBits int
	// DataBits is the number of payload data bits before the payload CRC.
	DataBits int
	// DataBytes is the number of payload bytes available to callers.
	DataBytes int
	// CodeOrder is log2 of the encoded payload bit count.
	CodeOrder int
	// SymbolCount is the number of payload OFDM symbols, excluding metadata.
	SymbolCount int
	// Duration is the approximate frame duration in seconds.
	Duration float64
	// BitrateKbps is the approximate user payload rate in kilobits per second.
	BitrateKbps float64
}

// Setup expands a Mode into all derived parameters needed by encoders and
// decoders. Most higher-level functions call Setup internally, but it is useful
// when you need to size payloads before encoding.
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
	frozenBits, err := FrozenPayloadBits(cfg.CodeOrder, cfg.CodeRate)
	if err != nil {
		return Config{}, err
	}
	cfg.DataBits = dataBits
	cfg.DataBytes = dataBits / 8
	cfg.FrozenBits = frozenBits
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
