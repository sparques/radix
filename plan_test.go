package radix

import "testing"

func TestBuildTonePlanUsesExpectedCapacity(t *testing.T) {
	for _, mod := range []Modulation{BPSK, QPSK, PSK8, QAM16, QAM64, QAM256, QAM1024, QAM4096} {
		for _, rate := range []CodeRate{RateHalf, RateTwoThirds, RateThreeQuarters, RateFiveSixths} {
			for _, frame := range []FrameSize{ShortFrame, NormalFrame} {
				mode, err := NewMode(mod, rate, frame)
				if err != nil {
					t.Fatalf("NewMode(%s, %s, %s): %v", mod, rate, frame, err)
				}
				cfg, err := Setup(mode)
				if err != nil {
					t.Fatalf("Setup(%s, %s, %s): %v", mod, rate, frame, err)
				}
				plans, err := BuildTonePlan(cfg)
				if err != nil {
					t.Fatalf("BuildTonePlan(%s, %s, %s): %v", mod, rate, frame, err)
				}
				if len(plans) != cfg.SymbolCount+1 {
					t.Fatalf("%s %s %s plans = %d, want %d", mod, rate, frame, len(plans), cfg.SymbolCount+1)
				}

				var metaBits, dataBits, seedTones int
				for _, plan := range plans {
					if len(plan.Tones) != ToneCount {
						t.Fatalf("%s %s %s symbol %d tones = %d, want %d", mod, rate, frame, plan.Index, len(plan.Tones), ToneCount)
					}
					for _, tone := range plan.Tones {
						switch tone.Kind {
						case SeedTone:
							seedTones++
						case MetaTone:
							metaBits += tone.Bits
						case DataTone:
							dataBits += tone.Bits
						}
					}
				}

				if metaBits != DataTones {
					t.Errorf("%s %s %s meta bits = %d, want %d", mod, rate, frame, metaBits, DataTones)
				}
				if dataBits != 1<<cfg.CodeOrder {
					t.Errorf("%s %s %s data bits = %d, want %d", mod, rate, frame, dataBits, 1<<cfg.CodeOrder)
				}
				wantSeedTones := SeedTones * (cfg.SymbolCount + 1)
				if seedTones != wantSeedTones {
					t.Errorf("%s %s %s seed tones = %d, want %d", mod, rate, frame, seedTones, wantSeedTones)
				}
			}
		}
	}
}

func TestTonePlanAdaptiveModulationBoundaries(t *testing.T) {
	tests := []struct {
		modBits int
		offset  int
		want    int
	}{
		{3, 27, 3},
		{3, 30, 2},
		{6, 54, 6},
		{6, 60, 4},
		{10, 110, 10},
		{10, 120, 8},
		{12, 108, 12},
		{12, 120, 8},
	}

	for _, tt := range tests {
		if got := dataToneBits(tt.modBits, tt.offset); got != tt.want {
			t.Errorf("dataToneBits(%d, %d) = %d, want %d", tt.modBits, tt.offset, got, tt.want)
		}
	}
}

func TestValidateFrequencyOffset(t *testing.T) {
	if err := ValidateFrequencyOffset(48000, 1, 1500); err != nil {
		t.Fatalf("ValidateFrequencyOffset accepted upstream quick-start offset: %v", err)
	}
	if err := ValidateFrequencyOffset(48000, 1, 1200); err != nil {
		t.Fatalf("ValidateFrequencyOffset accepted lower mono edge: %v", err)
	}
	if err := ValidateFrequencyOffset(48000, 1, 900); err == nil {
		t.Fatal("ValidateFrequencyOffset accepted mono offset below bandwidth edge")
	}
	if err := ValidateFrequencyOffset(48000, 2, -22800); err != nil {
		t.Fatalf("ValidateFrequencyOffset rejected stereo lower edge: %v", err)
	}
	if err := ValidateFrequencyOffset(48000, 2, 22900); err == nil {
		t.Fatal("ValidateFrequencyOffset accepted offset that is not divisible by 300")
	}
	if err := ValidateFrequencyOffset(48000, 2, 23100); err == nil {
		t.Fatal("ValidateFrequencyOffset accepted offset above upper edge")
	}
}

func TestToneOffset(t *testing.T) {
	got, err := ToneOffset(48000, 1500)
	if err != nil {
		t.Fatal(err)
	}
	if got != 40 {
		t.Errorf("ToneOffset(48000, 1500) = %d, want 40", got)
	}
}
