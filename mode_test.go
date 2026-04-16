package radix

import (
	"math"
	"testing"
)

func TestSetupMatchesUpstreamModeTable(t *testing.T) {
	tests := []struct {
		mod          Modulation
		rate         CodeRate
		frame        FrameSize
		payloadBytes int
		duration     float64
	}{
		{BPSK, RateHalf, ShortFrame, 128, 1.5033333333333334},
		{QPSK, RateHalf, NormalFrame, 512, 2.5966666666666667},
		{PSK8, RateTwoThirds, ShortFrame, 684, 1.9133333333333333},
		{QAM16, RateThreeQuarters, ShortFrame, 384, 0.9566666666666667},
		{QAM64, RateFiveSixths, NormalFrame, 3408, 3.4166666666666665},
		{QAM256, RateHalf, ShortFrame, 1024, 1.5033333333333334},
		{QAM1024, RateThreeQuarters, NormalFrame, 6144, 3.9633333333333334},
		{QAM4096, RateFiveSixths, NormalFrame, 6816, 3.4166666666666665},
	}

	for _, tt := range tests {
		mode, err := NewMode(tt.mod, tt.rate, tt.frame)
		if err != nil {
			t.Fatalf("NewMode(%s, %s, %s): %v", tt.mod, tt.rate, tt.frame, err)
		}
		cfg, err := Setup(mode)
		if err != nil {
			t.Fatalf("Setup(%s, %s, %s): %v", tt.mod, tt.rate, tt.frame, err)
		}
		if cfg.DataBytes != tt.payloadBytes {
			t.Errorf("%s %s %s payload = %d, want %d", tt.mod, tt.rate, tt.frame, cfg.DataBytes, tt.payloadBytes)
		}
		if math.Abs(cfg.Duration-tt.duration) > 1e-12 {
			t.Errorf("%s %s %s duration = %.15f, want %.15f", tt.mod, tt.rate, tt.frame, cfg.Duration, tt.duration)
		}
	}
}

func TestSetupRejectsUnsupportedModes(t *testing.T) {
	if _, err := Setup(Mode(128)); err == nil {
		t.Fatal("Setup accepted analog mode")
	}
	if _, err := Setup(Mode(uint8(QAM16)<<4 | 4<<1)); err == nil {
		t.Fatal("Setup accepted unsupported code rate")
	}
}
